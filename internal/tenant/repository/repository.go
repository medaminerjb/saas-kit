package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medaminerjb/saas-kit/internal/sqlcgen"
	"github.com/medaminerjb/saas-kit/internal/tenant/domain"
)

// TenantRepository defines the data access contract for tenants, members, and invitations.
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, []domain.MemberRole, error)

	AddMember(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole) error
	GetMember(ctx context.Context, tenantID, userID uuid.UUID) (*domain.Member, error)
	ListMembers(ctx context.Context, tenantID uuid.UUID) ([]*domain.Member, error)
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole) error

	CreateInvitation(ctx context.Context, invite *domain.Invitation) error
	GetInvitationByID(ctx context.Context, id uuid.UUID) (*domain.Invitation, error)
	GetInvitationByToken(ctx context.Context, tokenHash string) (*domain.Invitation, error)
	ListInvitations(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invitation, error)
	AcceptInvitation(ctx context.Context, id uuid.UUID) error
	DeleteInvitation(ctx context.Context, id uuid.UUID) error

	UpdateUserActiveTenant(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) error
}

type pgTenantRepo struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

// NewTenantRepository creates a new PostgreSQL tenant repository.
func NewTenantRepository(pool *pgxpool.Pool) TenantRepository {
	return &pgTenantRepo{
		pool:    pool,
		queries: sqlcgen.New(pool),
	}
}

func (r *pgTenantRepo) Create(ctx context.Context, t *domain.Tenant) error {
	res, err := r.queries.CreateTenant(ctx, sqlcgen.CreateTenantParams{
		Name:   t.Name,
		Slug:   t.Slug,
		Status: string(t.Status),
		Plan:   t.Plan,
	})
	if err != nil {
		return err
	}
	t.ID = res.ID
	t.CreatedAt = res.CreatedAt
	t.UpdatedAt = res.UpdatedAt
	return nil
}

func (r *pgTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	res, err := r.queries.GetTenantByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, err
	}
	return &domain.Tenant{
		ID:        res.ID,
		Name:      res.Name,
		Slug:      res.Slug,
		Status:    domain.TenantStatus(res.Status),
		Plan:      res.Plan,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}, nil
}

func (r *pgTenantRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	res, err := r.queries.GetTenantBySlug(ctx, slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, err
	}
	return &domain.Tenant{
		ID:        res.ID,
		Name:      res.Name,
		Slug:      res.Slug,
		Status:    domain.TenantStatus(res.Status),
		Plan:      res.Plan,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}, nil
}

func (r *pgTenantRepo) Update(ctx context.Context, t *domain.Tenant) error {
	res, err := r.queries.UpdateTenant(ctx, sqlcgen.UpdateTenantParams{
		ID:     t.ID,
		Name:   t.Name,
		Slug:   t.Slug,
		Status: string(t.Status),
		Plan:   t.Plan,
	})
	if err != nil {
		return err
	}
	t.UpdatedAt = res.UpdatedAt
	return nil
}

func (r *pgTenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteTenant(ctx, id)
}

func (r *pgTenantRepo) ListForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, []domain.MemberRole, error) {
	rows, err := r.queries.ListTenantsForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	tenants := make([]*domain.Tenant, len(rows))
	roles := make([]domain.MemberRole, len(rows))
	for i, row := range rows {
		tenants[i] = &domain.Tenant{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			Status:    domain.TenantStatus(row.Status),
			Plan:      row.Plan,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
		roles[i] = domain.MemberRole(row.Role)
	}
	return tenants, roles, nil
}

func (r *pgTenantRepo) AddMember(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole) error {
	_, err := r.queries.AddTenantMember(ctx, sqlcgen.AddTenantMemberParams{
		TenantID: tenantID,
		UserID:   userID,
		Role:     string(role),
	})
	return err
}

func (r *pgTenantRepo) GetMember(ctx context.Context, tenantID, userID uuid.UUID) (*domain.Member, error) {
	res, err := r.queries.GetTenantMember(ctx, sqlcgen.GetTenantMemberParams{
		TenantID: tenantID,
		UserID:   userID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("member not found")
		}
		return nil, err
	}
	return &domain.Member{
		TenantID: res.TenantID,
		UserID:   res.UserID,
		Role:     domain.MemberRole(res.Role),
		JoinedAt: res.JoinedAt,
	}, nil
}

func (r *pgTenantRepo) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]*domain.Member, error) {
	rows, err := r.queries.ListTenantMembers(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	members := make([]*domain.Member, len(rows))
	for i, row := range rows {
		members[i] = &domain.Member{
			TenantID:  row.TenantID,
			UserID:    row.UserID,
			Role:      domain.MemberRole(row.Role),
			JoinedAt:  row.JoinedAt,
			Email:     row.Email,
			Name:      row.Name,
			AvatarURL: row.AvatarUrl,
		}
	}
	return members, nil
}

func (r *pgTenantRepo) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.queries.RemoveTenantMember(ctx, sqlcgen.RemoveTenantMemberParams{
		TenantID: tenantID,
		UserID:   userID,
	})
}

func (r *pgTenantRepo) UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole) error {
	_, err := r.queries.UpdateTenantMemberRole(ctx, sqlcgen.UpdateTenantMemberRoleParams{
		TenantID: tenantID,
		UserID:   userID,
		Role:     string(role),
	})
	return err
}

func (r *pgTenantRepo) CreateInvitation(ctx context.Context, invite *domain.Invitation) error {
	res, err := r.queries.CreateTenantInvitation(ctx, sqlcgen.CreateTenantInvitationParams{
		TenantID:  invite.TenantID,
		Email:     invite.Email,
		Role:      string(invite.Role),
		TokenHash: invite.TokenHash,
		ExpiresAt: invite.ExpiresAt,
	})
	if err != nil {
		return err
	}
	invite.ID = res.ID
	invite.CreatedAt = res.CreatedAt
	return nil
}

func (r *pgTenantRepo) GetInvitationByID(ctx context.Context, id uuid.UUID) (*domain.Invitation, error) {
	res, err := r.queries.GetTenantInvitationByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, err
	}
	var acceptedAt *time.Time
	if res.AcceptedAt.Valid {
		acceptedAt = &res.AcceptedAt.Time
	}
	return &domain.Invitation{
		ID:         res.ID,
		TenantID:   res.TenantID,
		Email:      res.Email,
		Role:       domain.MemberRole(res.Role),
		TokenHash:  res.TokenHash,
		ExpiresAt:  res.ExpiresAt,
		AcceptedAt: acceptedAt,
		CreatedAt:  res.CreatedAt,
	}, nil
}

func (r *pgTenantRepo) GetInvitationByToken(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
	res, err := r.queries.GetTenantInvitationByToken(ctx, tokenHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invitation not found or expired")
		}
		return nil, err
	}
	var acceptedAt *time.Time
	if res.AcceptedAt.Valid {
		acceptedAt = &res.AcceptedAt.Time
	}
	return &domain.Invitation{
		ID:         res.ID,
		TenantID:   res.TenantID,
		Email:      res.Email,
		Role:       domain.MemberRole(res.Role),
		TokenHash:  res.TokenHash,
		ExpiresAt:  res.ExpiresAt,
		AcceptedAt: acceptedAt,
		CreatedAt:  res.CreatedAt,
	}, nil
}

func (r *pgTenantRepo) ListInvitations(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invitation, error) {
	rows, err := r.queries.ListTenantInvitations(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	invitations := make([]*domain.Invitation, len(rows))
	for i, row := range rows {
		var acceptedAt *time.Time
		if row.AcceptedAt.Valid {
			acceptedAt = &row.AcceptedAt.Time
		}
		invitations[i] = &domain.Invitation{
			ID:         row.ID,
			TenantID:   row.TenantID,
			Email:      row.Email,
			Role:       domain.MemberRole(row.Role),
			TokenHash:  row.TokenHash,
			ExpiresAt:  row.ExpiresAt,
			AcceptedAt: acceptedAt,
			CreatedAt:  row.CreatedAt,
		}
	}
	return invitations, nil
}

func (r *pgTenantRepo) AcceptInvitation(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.AcceptTenantInvitation(ctx, id)
	return err
}

func (r *pgTenantRepo) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteTenantInvitation(ctx, id)
}

func (r *pgTenantRepo) UpdateUserActiveTenant(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) error {
	var val pgtype.UUID
	if tenantID != nil {
		val = pgtype.UUID{Bytes: *tenantID, Valid: true}
	} else {
		val = pgtype.UUID{Valid: false}
	}
	_, err := r.queries.UpdateUserActiveTenant(ctx, sqlcgen.UpdateUserActiveTenantParams{
		ID:       userID,
		TenantID: val,
	})
	return err
}
