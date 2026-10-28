package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"kallpa-api/internal/domain"
)

var (
	// ErrInvalidCredentials indicates that login email or password are incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound indicates that the specified user could not be found.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidToken indicates that a provided token is missing, malformed or expired.
	ErrInvalidToken = errors.New("invalid token")
)

// TokenResponse contains both access and refresh JWT tokens.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// CustomClaims embeds standard claims alongside application specific ones.
type CustomClaims struct {
	UserID  int64           `json:"user_id"`
	StoreID *int64          `json:"store_id"`
	Role    domain.UserRole `json:"role"`
	Type    string          `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// AuthService outlines operations regarding authentication.
type AuthService interface {
	Login(ctx context.Context, email, password string) (*TokenResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error)
	ValidateToken(tokenStr string) (*CustomClaims, error)
	HashPassword(password string) (string, error)
}

type service struct {
	repo      UserRepository
	jwtSecret string
}

// NewService creates a new authentication service.
func NewService(repo UserRepository, jwtSecret string) AuthService {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// We can choose to rotate refresh token or just issue new access token.
	// We'll rotate both for better security.
	return s.generateTokens(user)
}

func (s *service) ValidateToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (s *service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *service) generateTokens(user *domain.User) (*TokenResponse, error) {
	// Access Token (e.g. 15 minutes)
	accessClaims := CustomClaims{
		UserID:  user.ID,
		StoreID: user.StoreID,
		Role:    user.Role,
		Type:    "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	// Refresh Token (e.g. 7 days)
	refreshClaims := CustomClaims{
		UserID:  user.ID,
		StoreID: user.StoreID,
		Role:    user.Role,
		Type:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}
