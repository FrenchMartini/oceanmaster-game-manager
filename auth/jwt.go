// Package auth provides JWT authentication functionality.
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when a token is invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when a token has expired.
	ErrExpiredToken = errors.New("token has expired")
)

// JWTService handles JWT operations
type JWTService struct {
	secret     string
	expiryTime time.Duration
}

// Claims represents JWT claims
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
func NewJWTService(secret string, expiryTime time.Duration) *JWTService {
	return &JWTService{
		secret:     secret,
		expiryTime: expiryTime,
	}
}

// GenerateToken generates a JWT token for a user
func (s *JWTService) GenerateToken(userID int64, email string) (string, error) {
	expirationTime := time.Now().Add(s.expiryTime)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// ValidateToken validates a JWT token and returns claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateStateToken generates a random state token for OAuth CSRF protection
func GenerateStateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		fmt.Println("warning: failed to generate state token:", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
