// Package models defines data models used throughout the application.
package models

import "time"

// User represents a user in the system
type User struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	GoogleID  string    `json:"-" db:"google_id"`
	Name      string    `json:"name" db:"name"`
	Picture   string    `json:"picture" db:"picture"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LoginResponse represents the login response with JWT token
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
