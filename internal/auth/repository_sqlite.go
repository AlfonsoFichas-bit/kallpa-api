package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"kallpa-api/internal/domain"
)

type UserPGRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserPGRepository {
	return &UserPGRepository{db: db}
}

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

func (r *UserPGRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
