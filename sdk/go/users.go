package saaskit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// UsersClient handles user management operations.
type UsersClient struct {
	client *Client
}

// NewUsersClient creates a new users client.
func NewUsersClient(client *Client) *UsersClient {
	return &UsersClient{client: client}
}

// User represents a user.
type User struct {
	ID             string                 `json:"id"`
	Email          string                 `json:"email"`
	FirstName      string                 `json:"first_name,omitempty"`
	LastName       string                 `json:"last_name,omitempty"`
	EmailVerified  bool                   `json:"email_verified"`
	Status         string                 `json:"status"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	MetadataPublic map[string]interface{} `json:"metadata_public,omitempty"`
}

// GetMe retrieves the current authenticated user.
func (c *UsersClient) GetMe(ctx context.Context, accessToken string) (*User, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.client.BaseURL()+"/api/v1/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user failed with status %d", resp.StatusCode)
	}

	var result User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UpdateMe updates the current authenticated user's profile.
func (c *UsersClient) UpdateMe(ctx context.Context, accessToken string, req *UpdateUserRequest) (*User, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.client.BaseURL()+"/api/v1/users/me", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update user failed with status %d", resp.StatusCode)
	}

	var result User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Session represents a user session.
type Session struct {
	ID           string    `json:"id"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	LastActiveAt string    `json:"last_active_at"`
	CreatedAt    string    `json:"created_at"`
	ExpiresAt    string    `json:"expires_at"`
}

// ListSessions retrieves the current user's sessions.
func (c *UsersClient) ListSessions(ctx context.Context, accessToken string) ([]Session, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.client.BaseURL()+"/api/v1/users/me/sessions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list sessions failed with status %d", resp.StatusCode)
	}

	var result struct {
		Sessions []Session `json:"sessions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Sessions, nil
}

// RevokeSession revokes a specific session.
func (c *UsersClient) RevokeSession(ctx context.Context, accessToken string, sessionID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.client.BaseURL()+"/api/v1/users/me/sessions/"+sessionID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(ctx, httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke session failed with status %d", resp.StatusCode)
	}

	return nil
}
