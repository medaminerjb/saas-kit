package saaskit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MetadataClient handles metadata operations.
type MetadataClient struct {
	client *Client
}

// NewMetadataClient creates a new metadata client.
func NewMetadataClient(client *Client) *MetadataClient {
	return &MetadataClient{client: client}
}

// UserMetadata represents user metadata.
type UserMetadata struct {
	MetadataPublic  map[string]interface{} `json:"metadata_public,omitempty"`
	MetadataPrivate map[string]interface{} `json:"metadata_private,omitempty"`
}

// GetUserMetadata retrieves the current user's metadata.
func (c *MetadataClient) GetUserMetadata(ctx context.Context, accessToken string) (*UserMetadata, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.client.BaseURL()+"/api/v1/users/me/metadata", nil)
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
		return nil, fmt.Errorf("get user metadata failed with status %d", resp.StatusCode)
	}

	var result UserMetadata
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateUserMetadataRequest represents a user metadata update request.
type UpdateUserMetadataRequest struct {
	MetadataPublic  map[string]interface{} `json:"metadata_public,omitempty"`
	MetadataPrivate map[string]interface{} `json:"metadata_private,omitempty"`
}

// UpdateUserMetadata updates the current user's metadata.
func (c *MetadataClient) UpdateUserMetadata(ctx context.Context, accessToken string, req *UpdateUserMetadataRequest) (*UserMetadata, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.client.BaseURL()+"/api/v1/users/me/metadata", bytes.NewReader(body))
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
		return nil, fmt.Errorf("update user metadata failed with status %d", resp.StatusCode)
	}

	var result UserMetadata
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// TenantMetadata represents tenant metadata.
type TenantMetadata struct {
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GetTenantMetadata retrieves a tenant's metadata.
func (c *MetadataClient) GetTenantMetadata(ctx context.Context, accessToken string, tenantID string) (*TenantMetadata, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.client.BaseURL()+"/api/v1/tenants/"+tenantID+"/metadata", nil)
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
		return nil, fmt.Errorf("get tenant metadata failed with status %d", resp.StatusCode)
	}

	var result TenantMetadata
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateTenantMetadataRequest represents a tenant metadata update request.
type UpdateTenantMetadataRequest struct {
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTenantMetadata updates a tenant's metadata.
func (c *MetadataClient) UpdateTenantMetadata(ctx context.Context, accessToken string, tenantID string, req *UpdateTenantMetadataRequest) (*TenantMetadata, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.client.BaseURL()+"/api/v1/tenants/"+tenantID+"/metadata", bytes.NewReader(body))
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
		return nil, fmt.Errorf("update tenant metadata failed with status %d", resp.StatusCode)
	}

	var result TenantMetadata
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
