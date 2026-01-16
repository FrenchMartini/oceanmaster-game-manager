// Package handlers provides HTTP request handlers.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "game-manager",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("warning: failed to encode response: %v\n", err)
	}
}
