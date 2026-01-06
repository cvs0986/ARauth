package hydra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Hydra admin API client
type Client struct {
	adminURL string
	httpClient *http.Client
}

// NewClient creates a new Hydra client
func NewClient(adminURL string) *Client {
	return &Client{
		adminURL: adminURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// AcceptLoginRequest represents a request to accept a login
type AcceptLoginRequest struct {
	Subject string                 `json:"subject"`
	Context map[string]interface{} `json:"context,omitempty"`
	Remember bool                  `json:"remember,omitempty"`
	RememberFor int                `json:"remember_for,omitempty"`
}

// AcceptLoginResponse represents the response from accepting a login
type AcceptLoginResponse struct {
	RedirectTo string `json:"redirect_to"`
}

// GetLoginRequest represents a login request from Hydra
type GetLoginRequest struct {
	Challenge       string                 `json:"challenge"`
	RequestedScope  []string               `json:"requested_scope"`
	RequestedAudience []string             `json:"requested_audience"`
	Skip            bool                   `json:"skip"`
	Subject         string                 `json:"subject"`
	Client          ClientInfo             `json:"client"`
	RequestURL      string                 `json:"request_url"`
	SessionID       string                 `json:"session_id"`
}

// ClientInfo represents OAuth2 client information
type ClientInfo struct {
	ClientID string `json:"client_id"`
}

// AcceptLoginRequest accepts a login request in Hydra
func (c *Client) AcceptLoginRequest(ctx context.Context, challenge string, req *AcceptLoginRequest) (*AcceptLoginResponse, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login/accept?login_challenge=%s", c.adminURL, challenge)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hydra error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var acceptResp AcceptLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&acceptResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &acceptResp, nil
}

// GetLoginRequest retrieves login request information from Hydra
func (c *Client) GetLoginRequest(ctx context.Context, challenge string) (*GetLoginRequest, error) {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login?login_challenge=%s", c.adminURL, challenge)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hydra error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginReq GetLoginRequest
	if err := json.NewDecoder(resp.Body).Decode(&loginReq); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &loginReq, nil
}

// RejectLoginRequest rejects a login request in Hydra
func (c *Client) RejectLoginRequest(ctx context.Context, challenge string, errorCode string, errorDescription string) error {
	url := fmt.Sprintf("%s/admin/oauth2/auth/requests/login/reject?login_challenge=%s", c.adminURL, challenge)

	reqBody := map[string]string{
		"error":             errorCode,
		"error_description": errorDescription,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hydra error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

