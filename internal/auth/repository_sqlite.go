package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"kallpa-api/internal/domain"
)

// UserPGRepository handles data access using GORM.
type UserPGRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a new instance of UserPGRepository.
func NewUserRepository(db *gorm.DB) *UserPGRepository {
	return &UserPGRepository{db: db}
}

// FindByEmail retrieves a user by their email address.
func (r *UserPGRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil when not found
		}
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by their primary key.
func (r *UserPGRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user record into the database.
func (r *UserPGRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
