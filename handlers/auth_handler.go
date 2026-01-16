// Package handlers provides HTTP request handlers.
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/oceanmining/game-manager/auth"
	"github.com/oceanmining/game-manager/config"
	"github.com/oceanmining/game-manager/models"
)

var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when a user already exists.
	ErrUserExists = errors.New("user already exists")
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db          *sql.DB
	jwtService  *auth.JWTService
	oauthConfig *oauth2.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *sql.DB, jwtService *auth.JWTService, googleConfig config.GoogleOAuthConfig) *AuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     googleConfig.ClientID,
		ClientSecret: googleConfig.ClientSecret,
		RedirectURL:  googleConfig.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthHandler{
		db:          db,
		jwtService:  jwtService,
		oauthConfig: oauthConfig,
	}
}

// GoogleLogin initiates Google OAuth flow
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig.ClientID == "" || h.oauthConfig.ClientSecret == "" {
		http.Error(w, "Google OAuth not configured", http.StatusInternalServerError)
		return
	}

	// Generate state token for CSRF protection
	state := auth.GenerateStateToken()

	// Store state in session/cookie (simplified - in production use secure cookies)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state token
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("warning: failed to close response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		http.Error(w, "Failed to parse user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find or create user
	user, err := h.findOrCreateUser(googleUser.ID, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		http.Error(w, "Failed to create/find user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	jwtToken, err := h.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return token and user info
	response := models.LoginResponse{
		Token: jwtToken,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("warning: failed to encode response: %v\n", err)
	}
}

// findOrCreateUser finds an existing user or creates a new one
func (h *AuthHandler) findOrCreateUser(googleID, email, name, picture string) (*models.User, error) {
	var user models.User

	// Try to find existing user by Google ID
	query := `SELECT id, email, google_id, name, picture, created_at FROM users WHERE google_id = $1`
	err := h.db.QueryRow(query, googleID).Scan(
		&user.ID,
		&user.Email,
		&user.GoogleID,
		&user.Name,
		&user.Picture,
		&user.CreatedAt,
	)

	if err == nil {
		// User exists, return it
		return &user, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// User doesn't exist, create new one
	query = `INSERT INTO users (email, google_id, name, picture) VALUES ($1, $2, $3, $4) RETURNING id, email, google_id, name, picture, created_at`
	err = h.db.QueryRow(query, email, googleID, name, picture).Scan(
		&user.ID,
		&user.Email,
		&user.GoogleID,
		&user.Name,
		&user.Picture,
		&user.CreatedAt,
	)

	if err != nil {
		// Check if it's a duplicate email error
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			// Try to find by email instead
			query = `SELECT id, email, google_id, name, picture, created_at FROM users WHERE email = $1`
			err = h.db.QueryRow(query, email).Scan(
				&user.ID,
				&user.Email,
				&user.GoogleID,
				&user.Name,
				&user.Picture,
				&user.CreatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to find user: %w", err)
			}
			return &user, nil
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}
