package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"kallpa-api/internal/domain"
)

type mockUserRepo struct {
	user *domain.User
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.user != nil && m.user.Email == email {
		return m.user, nil
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.user != nil && m.user.ID == id {
		return m.user, nil
	}
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.user = user
	return nil
}

func TestAuthService_HashPassword(t *testing.T) {
	svc := NewService(&mockUserRepo{}, "secret")

	pwd := "password123"
	hash, err := svc.HashPassword(pwd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hash == pwd {
		t.Errorf("expected hash to be different from password")
	}
}

func TestAuthService_Login(t *testing.T) {
	svc := NewService(nil, "secret")
	hash, _ := svc.HashPassword("securepass")

	repo := &mockUserRepo{
		user: &domain.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: hash,
			Role:         domain.RoleStoreOwner,
		},
	}
	svc = NewService(repo, "secret")

	// Test successful login
	tokens, err := svc.Login(context.Background(), "test@example.com", "securepass")
	if err != nil {
		t.Fatalf("expected no error on valid login, got %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("expected both tokens to be generated")
	}

	// Test invalid password
	_, err = svc.Login(context.Background(), "test@example.com", "wrongpass")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials on wrong password, got %v", err)
	}

	// Test invalid email
	_, err = svc.Login(context.Background(), "wrong@example.com", "securepass")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials on wrong email, got %v", err)
	}
}

func TestAuthService_RefreshAndValidate(t *testing.T) {
	repo := &mockUserRepo{
		user: &domain.User{
			ID:    1,
			Email: "test@example.com",
			Role:  domain.RoleSalesAgent,
		},
	}
	svc := NewService(repo, "secret")
	svcImpl := svc.(*service)

	tokens, err := svcImpl.generateTokens(repo.user)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	// Validate access token
	claims, err := svc.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("expected access token to be valid, got %v", err)
	}
	if claims.Type != "access" || claims.UserID != 1 || claims.Role != domain.RoleSalesAgent {
		t.Errorf("invalid access token claims: %+v", claims)
	}

	// Test successful refresh
	newTokens, err := svc.Refresh(context.Background(), tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected refresh to succeed, got %v", err)
	}
	if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
		t.Errorf("expected new tokens")
	}

	// Test refresh with access token (should fail)
	_, err = svc.Refresh(context.Background(), tokens.AccessToken)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken when refreshing with access token, got %v", err)
	}

	// Test expired token validation
	expiredClaims := CustomClaims{
		UserID: 1,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenStr, _ := expiredToken.SignedString([]byte("secret"))

	_, err = svc.ValidateToken(expiredTokenStr)
	if err == nil {
		t.Errorf("expected error for expired token")
	}
}
