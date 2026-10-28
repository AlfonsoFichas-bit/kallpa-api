package auth

import (
	"context"

	"kallpa-api/internal/domain"
)

// UserRepository specifies the database operations required for users.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}
