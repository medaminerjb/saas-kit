package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/platform/events"
	"github.com/medaminerjb/saas-kit/internal/tenant/domain"
	"github.com/medaminerjb/saas-kit/internal/tenant/repository"
)

// TenantService handles all business logic related to tenants, memberships, and invitations.
type TenantService struct {
	repo      repository.TenantRepository
	publisher events.Publisher
	logger    *slog.Logger
}

// NewTenantService creates a new tenant service instance.
func NewTenantService(repo repository.TenantRepository, publisher events.Publisher, logger *slog.Logger) *TenantService {
	return &TenantService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// CreateTenant creates a new organization and assigns the creator as the Owner.
func (s *TenantService) CreateTenant(ctx context.Context, name, slug string, userID uuid.UUID) (*domain.Tenant, error) {
	if name == "" {
		return nil, fmt.Errorf("tenant name is required")
	}

	if slug == "" {
		slug = slugify(name)
	} else {
		slug = slugify(slug)
	}

	if slug == "" {
		return nil, fmt.Errorf("invalid tenant name or slug")
	}

	// Check if slug is unique
	existing, _ := s.repo.GetBySlug(ctx, slug)
	if existing != nil {
		return nil, fmt.Errorf("organization slug already exists")
	}

	tenant := &domain.Tenant{
		Name:   name,
		Slug:   slug,
		Status: domain.TenantStatusActive,
		Plan:   "free",
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Add creator as Owner
	if err := s.repo.AddMember(ctx, tenant.ID, userID, domain.RoleOwner); err != nil {
		// Attempt cleanup of the created tenant on failure
		_ = s.repo.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to add owner membership: %w", err)
	}

	// Update user's active tenant to the new organization
	if err := s.repo.UpdateUserActiveTenant(ctx, userID, &tenant.ID); err != nil {
		s.logger.WarnContext(ctx, "failed to update user active tenant context after creation", "user_id", userID, "tenant_id", tenant.ID)
	}

	// Publish event
	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "tenant.created",
		ActorID:   &userID,
		TargetID:  &tenant.ID,
		Payload:   map[string]string{"name": tenant.Name, "slug": tenant.Slug},
		Timestamp: time.Now(),
	})

	s.logger.InfoContext(ctx, "tenant created successfully", "tenant_id", tenant.ID, "slug", tenant.Slug, "owner_id", userID)

	return tenant, nil
}

// GetTenant retrieves a tenant by ID.
func (s *TenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	return s.repo.GetByID(ctx, tenantID)
}

// GetTenantBySlug retrieves a tenant by Slug.
func (s *TenantService) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	return s.repo.GetBySlug(ctx, slugify(slug))
}

// UpdateTenant updates tenant settings.
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, name, slug string) (*domain.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if name != "" {
		tenant.Name = name
	}

	if slug != "" {
		cleanedSlug := slugify(slug)
		if cleanedSlug != tenant.Slug {
			existing, _ := s.repo.GetBySlug(ctx, cleanedSlug)
			if existing != nil {
				return nil, fmt.Errorf("organization slug already exists")
			}
			tenant.Slug = cleanedSlug
		}
	}

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// ListTenantsForUser returns all tenants the user is a member of.
func (s *TenantService) ListTenantsForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, []domain.MemberRole, error) {
	return s.repo.ListForUser(ctx, userID)
}

// InviteMember generates a secure invitation token and stores it in the database.
func (s *TenantService) InviteMember(ctx context.Context, tenantID uuid.UUID, email string, role domain.MemberRole) (*domain.Invitation, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, "", fmt.Errorf("email is required")
	}

	switch role {
	case domain.RoleOwner, domain.RoleAdmin, domain.RoleManager, domain.RoleMember, domain.RoleViewer:
		// allowed
	default:
		return nil, "", fmt.Errorf("invalid member role")
	}

	// Generate cryptographically secure raw token and its sha256 hash
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, "", fmt.Errorf("failed to generate secure invitation token: %w", err)
	}
	rawToken := hex.EncodeToString(b)
	h := sha256.New()
	h.Write([]byte(rawToken))
	tokenHash := hex.EncodeToString(h.Sum(nil))

	invite := &domain.Invitation{
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Valid for 24 hours
	}

	if err := s.repo.CreateInvitation(ctx, invite); err != nil {
		return nil, "", fmt.Errorf("failed to save invitation: %w", err)
	}

	s.logger.InfoContext(ctx, "member invited", "tenant_id", tenantID, "email", email, "role", role)

	return invite, rawToken, nil
}

// AcceptInvitation adds the user to the tenant and marks the invitation as accepted.
func (s *TenantService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	h := sha256.New()
	h.Write([]byte(token))
	tokenHash := hex.EncodeToString(h.Sum(nil))

	invite, err := s.repo.GetInvitationByToken(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("invalid or expired invitation token: %w", err)
	}

	if invite.IsExpired() {
		return fmt.Errorf("invitation has expired")
	}

	// Check if already a member
	existing, _ := s.repo.GetMember(ctx, invite.TenantID, userID)
	if existing != nil {
		// Update role if different
		if existing.Role != invite.Role {
			if err := s.repo.UpdateMemberRole(ctx, invite.TenantID, userID, invite.Role); err != nil {
				s.logger.WarnContext(ctx, "failed to update member role", "user_id", userID, "old_role", existing.Role, "new_role", invite.Role)
			}
		}
		// Update invitation status as accepted and return success
		_ = s.repo.AcceptInvitation(ctx, invite.ID)
		return nil
	}

	// Add member
	if err := s.repo.AddMember(ctx, invite.TenantID, userID, invite.Role); err != nil {
		return fmt.Errorf("failed to join tenant: %w", err)
	}

	// Accept invitation
	if err := s.repo.AcceptInvitation(ctx, invite.ID); err != nil {
		s.logger.WarnContext(ctx, "failed to mark invitation as accepted", "invitation_id", invite.ID)
	}

	// Set as user's active tenant
	_ = s.repo.UpdateUserActiveTenant(ctx, userID, &invite.TenantID)

	s.logger.InfoContext(ctx, "invitation accepted successfully", "tenant_id", invite.TenantID, "user_id", userID)

	return nil
}

// ListMembers lists all members of the tenant.
func (s *TenantService) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]*domain.Member, error) {
	return s.repo.ListMembers(ctx, tenantID)
}

// RemoveMember deletes a member from the tenant.
func (s *TenantService) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Verify membership exists
	m, err := s.repo.GetMember(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	if m.Role == domain.RoleOwner {
		return fmt.Errorf("cannot remove organization owner")
	}

	if err := s.repo.RemoveMember(ctx, tenantID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Reset active tenant for the user if they were removed from their active tenant
	// We check their active tenant context and if it matches the removed tenant, we set it to null
	// Wait, we don't have access to user details here directly, but we can call UpdateUserActiveTenant to nil
	// Or leave it to fall back on next login. Setting it to nil is cleaner.
	_ = s.repo.UpdateUserActiveTenant(ctx, userID, nil)

	s.logger.InfoContext(ctx, "member removed from tenant", "tenant_id", tenantID, "user_id", userID)

	return nil
}

// SwitchActiveTenant changes the active tenant context for the user.
func (s *TenantService) SwitchActiveTenant(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error {
	// Verify membership
	_, err := s.repo.GetMember(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("user is not a member of the target tenant")
	}

	if err := s.repo.UpdateUserActiveTenant(ctx, userID, &tenantID); err != nil {
		return fmt.Errorf("failed to switch tenant: %w", err)
	}

	s.logger.InfoContext(ctx, "user switched active tenant", "user_id", userID, "tenant_id", tenantID)

	return nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			sb.WriteRune('-')
		}
	}
	res := sb.String()
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	return strings.Trim(res, "-")
}
