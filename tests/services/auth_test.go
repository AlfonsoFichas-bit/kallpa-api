package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"kallpa-api/internal/auth"
	"kallpa-api/internal/domain"
)

const testJWTSecret = "supersecretkey"

func TestAuthService_Login(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := auth.NewService(mockRepo, testJWTSecret)

		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := &domain.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         domain.RoleStoreOwner,
		}

		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

		tokens, err := svc.Login(context.Background(), "test@example.com", password)
		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := auth.NewService(mockRepo, testJWTSecret)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("realpassword"), bcrypt.DefaultCost)
		user := &domain.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
		}

		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

		tokens, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
		assert.Nil(t, tokens)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := auth.NewService(mockRepo, testJWTSecret)

		mockRepo.On("FindByEmail", mock.Anything, "unknown@example.com").Return(nil, nil)

		tokens, err := svc.Login(context.Background(), "unknown@example.com", "password")
		assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
		assert.Nil(t, tokens)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := auth.NewService(mockRepo, testJWTSecret)

		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("db error"))

		tokens, err := svc.Login(context.Background(), "test@example.com", "password")
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		assert.Nil(t, tokens)
	})
}

func TestAuthService_Refresh(t *testing.T) {
	// Setup a valid refresh token using the service itself
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo, testJWTSecret)

	// Create token by manually generating it (bypassing Login DB call)
	// We'll reuse the service logic by making a valid login

	t.Run("Success", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := &domain.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         domain.RoleStoreOwner,
		}

		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
		tokens, err := svc.Login(context.Background(), "test@example.com", password)
		assert.NoError(t, err)

		// Now test Refresh
		mockRepo.On("FindByID", mock.Anything, int64(1)).Return(user, nil)

		newTokens, err := svc.Refresh(context.Background(), tokens.RefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, newTokens)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)
	})

	t.Run("Invalid Token - Used Access Token for Refresh", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := &domain.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         domain.RoleStoreOwner,
		}

		mockRepo2 := new(MockUserRepository)
		svc2 := auth.NewService(mockRepo2, testJWTSecret)
		mockRepo2.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
		tokens, _ := svc2.Login(context.Background(), "test@example.com", password)

		// Try to refresh with AccessToken instead of RefreshToken
		newTokens, err := svc2.Refresh(context.Background(), tokens.AccessToken)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, newTokens)
	})

	t.Run("User Not Found During Refresh", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := &domain.User{
			ID:           99,
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         domain.RoleStoreOwner,
		}

		mockRepo3 := new(MockUserRepository)
		svc3 := auth.NewService(mockRepo3, testJWTSecret)
		mockRepo3.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
		tokens, _ := svc3.Login(context.Background(), "test@example.com", password)

		// Mock user no longer exists
		mockRepo3.On("FindByID", mock.Anything, int64(99)).Return(nil, nil)

		newTokens, err := svc3.Refresh(context.Background(), tokens.RefreshToken)
		assert.ErrorIs(t, err, auth.ErrUserNotFound)
		assert.Nil(t, newTokens)
	})
}
