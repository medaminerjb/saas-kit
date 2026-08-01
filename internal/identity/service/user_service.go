package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/saaskit/saaskit/internal/identity/domain"
	"github.com/saaskit/saaskit/internal/identity/repository"
	"github.com/saaskit/saaskit/internal/platform/events"
)

// UserService handles user profile management operations.
type UserService struct {
	users     repository.UserRepository
	publisher events.Publisher
	logger    *slog.Logger
}

// NewUserService creates a new user management service.
func NewUserService(users repository.UserRepository, publisher events.Publisher, logger *slog.Logger) *UserService {
	return &UserService{
		users:     users,
		publisher: publisher,
		logger:    logger,
	}
}

// GetByID retrieves a user by their ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// UpdateProfileInput holds the fields that can be updated on a user profile.
type UpdateProfileInput struct {
	Name      *string
	AvatarURL *string
}

// UpdateProfile updates a user's profile information.
func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, input UpdateProfileInput) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:      "user.updated",
		ActorID:   &id,
		TargetID:  &id,
	})

	return user, nil
}

// DisableUser disables a user account.
func (s *UserService) DisableUser(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	user.Status = domain.UserStatusDisabled
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("disabling user: %w", err)
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:     "user.disabled",
		ActorID:  &actorID,
		TargetID: &id,
	})

	return nil
}

// DeleteUser performs a soft delete on a user account.
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	if err := s.users.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("soft deleting user: %w", err)
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:     "user.deleted",
		ActorID:  &actorID,
		TargetID: &id,
	})

	return nil
}

// ListUsers returns a paginated list of users.
func (s *UserService) ListUsers(ctx context.Context, tenantID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	users, err := s.users.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}

	total, err := s.users.Count(ctx, tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	return users, total, nil
}
