package saaskit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TenantsClient handles tenant/organization management operations.
type TenantsClient struct {
	client *Client
}

// NewTenantsClient creates a new tenants client.
func NewTenantsClient(client *Client) *TenantsClient {
	return &TenantsClient{client: client}
}

// Tenant represents a tenant/organization.
type Tenant struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Status    string                 `json:"status"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TenantWithRole represents a tenant with the user's role.
type TenantWithRole struct {
	Tenant *Tenant `json:"tenant"`
	Role   string  `json:"role"`
}

// CreateTenantRequest represents a tenant creation request.
type CreateTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// CreateTenant creates a new tenant/organization.
func (c *TenantsClient) Create(ctx context.Context, accessToken string, req *CreateTenantRequest) (*Tenant, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.client.BaseURL()+"/api/v1/tenants", bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create tenant failed with status %d", resp.StatusCode)
	}

	var result Tenant
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListTenants retrieves the current user's tenants.
func (c *TenantsClient) List(ctx context.Context, accessToken string) ([]TenantWithRole, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.client.BaseURL()+"/api/v1/tenants", nil)
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
		return nil, fmt.Errorf("list tenants failed with status %d", resp.StatusCode)
	}

	var result struct {
		Tenants []TenantWithRole `json:"tenants"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Tenants, nil
}

// GetTenant retrieves a specific tenant by ID.
func (c *TenantsClient) Get(ctx context.Context, accessToken string, tenantID string) (*Tenant, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.client.BaseURL()+"/api/v1/tenants/"+tenantID, nil)
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
		return nil, fmt.Errorf("get tenant failed with status %d", resp.StatusCode)
	}

	var result Tenant
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateTenantRequest represents a tenant update request.
type UpdateTenantRequest struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// UpdateTenant updates a tenant's information.
func (c *TenantsClient) Update(ctx context.Context, accessToken string, tenantID string, req *UpdateTenantRequest) (*Tenant, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.client.BaseURL()+"/api/v1/tenants/"+tenantID, bytes.NewReader(body))
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
		return nil, fmt.Errorf("update tenant failed with status %d", resp.StatusCode)
	}

	var result Tenant
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SwitchTenantRequest represents a tenant switch request.
type SwitchTenantRequest struct {
	TenantID string `json:"tenant_id"`
}

// SwitchTenant switches the user's active tenant.
func (c *TenantsClient) Switch(ctx context.Context, accessToken string, req *SwitchTenantRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.client.BaseURL()+"/api/v1/tenants/switch", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(ctx, httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("switch tenant failed with status %d", resp.StatusCode)
	}

	return nil
}

// AcceptInvitationRequest represents an invitation acceptance request.
type AcceptInvitationRequest struct {
	Token string `json:"token"`
}

// AcceptInvitation accepts a tenant invitation.
func (c *TenantsClient) AcceptInvitation(ctx context.Context, accessToken string, req *AcceptInvitationRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.client.BaseURL()+"/api/v1/tenants/invitations/accept", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(ctx, httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("accept invitation failed with status %d", resp.StatusCode)
	}

	return nil
}

// Member represents a tenant member.
type Member struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// ListMembers retrieves the members of a tenant.
func (c *TenantsClient) ListMembers(ctx context.Context, accessToken string, tenantID string) ([]Member, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.client.BaseURL()+"/api/v1/tenants/"+tenantID+"/members", nil)
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
		return nil, fmt.Errorf("list members failed with status %d", resp.StatusCode)
	}

	var result struct {
		Members []Member `json:"members"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Members, nil
}

// InviteMemberRequest represents a member invitation request.
type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// InviteMember invites a new member to a tenant.
func (c *TenantsClient) InviteMember(ctx context.Context, accessToken string, tenantID string, req *InviteMemberRequest) (*Member, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.client.BaseURL()+"/api/v1/tenants/"+tenantID+"/members", bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("invite member failed with status %d", resp.StatusCode)
	}

	var result Member
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// RemoveMember removes a member from a tenant.
func (c *TenantsClient) RemoveMember(ctx context.Context, accessToken string, tenantID string, userID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.client.BaseURL()+"/api/v1/tenants/"+tenantID+"/members/"+userID, nil)
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
		return fmt.Errorf("remove member failed with status %d", resp.StatusCode)
	}

	return nil
}
