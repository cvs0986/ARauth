package oauthclient

import (
	"time"

	"github.com/google/uuid"
)

// CreateClientRequest represents a request to create an OAuth2 client
type CreateClientRequest struct {
	Name           string   `json:"name" validate:"required,min=3,max=255"`
	Description    string   `json:"description" validate:"max=1000"`
	RedirectURIs   []string `json:"redirect_uris" validate:"required,min=1"`
	GrantTypes     []string `json:"grant_types" validate:"required,min=1"`
	Scopes         []string `json:"scopes" validate:"required,min=1"`
	IsConfidential bool     `json:"is_confidential"`
}

// CreateClientResponse includes the one-time secret (NEVER LOGGED, NEVER RE-SHOWN)
type CreateClientResponse struct {
	ID             uuid.UUID `json:"id"`
	ClientID       string    `json:"client_id"`
	ClientSecret   string    `json:"client_secret"` // ONE-TIME ONLY - never stored, never logged
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	RedirectURIs   []string  `json:"redirect_uris"`
	GrantTypes     []string  `json:"grant_types"`
	Scopes         []string  `json:"scopes"`
	IsConfidential bool      `json:"is_confidential"`
	CreatedAt      time.Time `json:"created_at"`
}

// Client represents an OAuth2 client (WITHOUT secret - safe for listing)
type Client struct {
	ID             uuid.UUID `json:"id"`
	ClientID       string    `json:"client_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	RedirectURIs   []string  `json:"redirect_uris"`
	GrantTypes     []string  `json:"grant_types"`
	Scopes         []string  `json:"scopes"`
	IsConfidential bool      `json:"is_confidential"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RotateSecretResponse includes the new one-time secret (NEVER LOGGED, NEVER RE-SHOWN)
type RotateSecretResponse struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"` // ONE-TIME ONLY - never stored, never logged
	RotatedAt    time.Time `json:"rotated_at"`
}
