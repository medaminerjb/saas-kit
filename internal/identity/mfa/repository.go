package mfa

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/medaminerjb/saas-kit/internal/sqlcgen"
)

var (
	ErrMethodNotFound = errors.New("mfa method not found")
	ErrNoMethods      = errors.New("no mfa methods configured")
)

// Repository defines the interface for MFA data operations.
type Repository interface {
	CreateMethod(ctx context.Context, method *Method) error
	GetMethodByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Method, error)
	GetMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*Method, error)
	GetDefaultMethod(ctx context.Context, userID uuid.UUID) (*Method, error)
	UpdateMethod(ctx context.Context, method *Method) error
	DeleteMethod(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	SetDefaultMethod(ctx context.Context, userID, methodID uuid.UUID) error

	CreateRecoveryCodes(ctx context.Context, codes []*RecoveryCode) error
	GetRecoveryCodesByUserID(ctx context.Context, userID uuid.UUID) ([]*RecoveryCode, error)
	MarkRecoveryCodeUsed(ctx context.Context, codeID uuid.UUID, userID uuid.UUID) error
}

type repository struct {
	queries *sqlcgen.Queries
}

// NewRepository creates a new MFA repository.
func NewRepository(queries *sqlcgen.Queries) Repository {
	return &repository{queries: queries}
}

func (r *repository) CreateMethod(ctx context.Context, method *Method) error {
	params := sqlcgen.CreateMFAMethodParams{
		UserID:    method.UserID,
		Type:      string(method.Type),
		Name:      &method.Name,
		SecretEnc: &method.SecretEnc,
		Verified:  method.Verified,
		IsDefault: method.IsDefault,
	}
	result, err := r.queries.CreateMFAMethod(ctx, params)
	if err != nil {
		return err
	}
	method.ID = result.ID
	method.CreatedAt = result.CreatedAt
	return nil
}

func (r *repository) GetMethodByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Method, error) {
	method, err := r.queries.GetMFAMethodByID(ctx, sqlcgen.GetMFAMethodByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMethodNotFound
		}
		return nil, err
	}
	return r.toDomainMethod(method), nil
}

func (r *repository) GetMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*Method, error) {
	methods, err := r.queries.ListMFAMethodsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*Method, len(methods))
	for i, m := range methods {
		result[i] = r.toDomainMethod(m)
	}
	return result, nil
}

func (r *repository) GetDefaultMethod(ctx context.Context, userID uuid.UUID) (*Method, error) {
	method, err := r.queries.GetVerifiedTOTPForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoMethods
		}
		return nil, err
	}
	return r.toDomainMethod(method), nil
}

func (r *repository) UpdateMethod(ctx context.Context, method *Method) error {
	// Use MarkMFAMethodVerified for verification updates
	if method.Verified {
		_, err := r.queries.MarkMFAMethodVerified(ctx, sqlcgen.MarkMFAMethodVerifiedParams{
			ID:     method.ID,
			UserID: method.UserID,
		})
		return err
	}
	return errors.New("only verification updates supported")
}

func (r *repository) DeleteMethod(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.queries.DeleteMFAMethod(ctx, sqlcgen.DeleteMFAMethodParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *repository) SetDefaultMethod(ctx context.Context, userID, methodID uuid.UUID) error {
	// Update is_default flag directly via UPDATE
	return errors.New("set default not implemented - use is_default on create")
}

func (r *repository) CreateRecoveryCodes(ctx context.Context, codes []*RecoveryCode) error {
	for _, code := range codes {
		err := r.queries.CreateMFARecoveryCode(ctx, sqlcgen.CreateMFARecoveryCodeParams{
			UserID:   code.UserID,
			CodeHash: code.CodeHash,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) GetRecoveryCodesByUserID(ctx context.Context, userID uuid.UUID) ([]*RecoveryCode, error) {
	codes, err := r.queries.ListUnusedRecoveryCodes(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*RecoveryCode, len(codes))
	for i, c := range codes {
		result[i] = &RecoveryCode{
			ID:        c.ID,
			UserID:    c.UserID,
			CodeHash:  c.CodeHash,
			CreatedAt: c.CreatedAt,
		}
		if c.UsedAt.Valid {
			result[i].UsedAt = &c.UsedAt.Time
		}
	}
	return result, nil
}

func (r *repository) MarkRecoveryCodeUsed(ctx context.Context, codeID uuid.UUID, userID uuid.UUID) error {
	return r.queries.MarkRecoveryCodeUsed(ctx, sqlcgen.MarkRecoveryCodeUsedParams{
		ID:     codeID,
		UserID: userID,
	})
}

func (r *repository) toDomainMethod(m sqlcgen.MfaMethod) *Method {
	method := &Method{
		ID:        m.ID,
		UserID:    m.UserID,
		Type:      MFAType(m.Type),
		Verified:  m.Verified,
		IsDefault: m.IsDefault,
		CreatedAt: m.CreatedAt,
	}
	if m.Name != nil {
		method.Name = *m.Name
	}
	if m.SecretEnc != nil {
		method.SecretEnc = *m.SecretEnc
	}
	if m.LastUsedAt.Valid {
		method.LastUsedAt = &m.LastUsedAt.Time
	}
	return method
}
