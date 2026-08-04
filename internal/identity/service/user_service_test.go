package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/platform/events"
)

// mockUserRepository implements repository.UserRepository for testing.
type mockUserRepository struct {
	getByIDFunc        func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	getByEmailFunc     func(ctx context.Context, email string, tenantID *uuid.UUID) (*domain.User, error)
	updateFunc         func(ctx context.Context, user *domain.User) error
	updateMetadataFunc func(ctx context.Context, id uuid.UUID, metadataPublic, metadataPrivate map[string]interface{}) error
	softDeleteFunc     func(ctx context.Context, id uuid.UUID) error
	listFunc           func(ctx context.Context, tenantID *uuid.UUID, limit, offset int32) ([]*domain.User, error)
	countFunc          func(ctx context.Context, tenantID *uuid.UUID) (int64, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string, tenantID *uuid.UUID) (*domain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email, tenantID)
	}
	return nil, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) UpdateMetadata(ctx context.Context, id uuid.UUID, metadataPublic, metadataPrivate map[string]interface{}) error {
	if m.updateMetadataFunc != nil {
		return m.updateMetadataFunc(ctx, id, metadataPublic, metadataPrivate)
	}
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil
}

func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockUserRepository) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.softDeleteFunc != nil {
		return m.softDeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, tenantID *uuid.UUID, limit, offset int32) ([]*domain.User, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, tenantID, limit, offset)
	}
	return nil, nil
}

func (m *mockUserRepository) Count(ctx context.Context, tenantID *uuid.UUID) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, tenantID)
	}
	return 0, nil
}

// mockPublisher implements events.Publisher for testing.
type mockPublisher struct {
	events []events.Event
}

func (m *mockPublisher) Publish(ctx context.Context, event events.Event) error {
	m.events = append(m.events, event)
	return nil
}

func TestUserService_GetByID(t *testing.T) {
	uID := uuid.New()
	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id == uID {
				return &domain.User{ID: uID, Name: "Test User"}, nil
			}
			return nil, errors.New("not found")
		},
	}
	pub := &mockPublisher{}
	us := NewUserService(mockRepo, pub, slog.Default())

	t.Run("success", func(t *testing.T) {
		user, err := us.GetByID(context.Background(), uID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Name != "Test User" {
			t.Errorf("expected name Test User, got %s", user.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := us.GetByID(context.Background(), uuid.New())
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	uID := uuid.New()
	existingUser := &domain.User{ID: uID, Name: "Old Name"}
	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return existingUser, nil
		},
		updateFunc: func(ctx context.Context, user *domain.User) error {
			existingUser = user
			return nil
		},
	}
	pub := &mockPublisher{}
	us := NewUserService(mockRepo, pub, slog.Default())

	newName := "New Name"
	_, err := us.UpdateProfile(context.Background(), uID, UpdateProfileInput{Name: &newName})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if existingUser.Name != "New Name" {
		t.Errorf("expected name to be updated to New Name, got %s", existingUser.Name)
	}

	if len(pub.events) != 1 || pub.events[0].Type != "user.updated" {
		t.Errorf("expected user.updated event, got: %v", pub.events)
	}
}

func TestUserService_DisableUser(t *testing.T) {
	uID := uuid.New()
	existingUser := &domain.User{ID: uID, Status: domain.UserStatusActive}
	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return existingUser, nil
		},
		updateFunc: func(ctx context.Context, user *domain.User) error {
			existingUser = user
			return nil
		},
	}
	pub := &mockPublisher{}
	us := NewUserService(mockRepo, pub, slog.Default())

	actorID := uuid.New()
	err := us.DisableUser(context.Background(), uID, actorID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if existingUser.Status != domain.UserStatusDisabled {
		t.Errorf("expected status to be disabled, got %s", existingUser.Status)
	}

	if len(pub.events) != 1 || pub.events[0].Type != "user.disabled" {
		t.Errorf("expected user.disabled event, got: %v", pub.events)
	}
}
