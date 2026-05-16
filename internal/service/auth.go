package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/username/parking-service/internal/config"
	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
	"github.com/username/parking-service/internal/validator"
)

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type AuthService struct {
	users userStore
	cfg   config.Config
}

type userStore interface {
	Create(context.Context, repository.CreateUserParams) (model.User, error)
	GetByEmail(context.Context, string) (model.User, error)
	GetByID(context.Context, uuid.UUID) (model.User, error)
	List(context.Context, int, int) ([]model.User, error)
	Update(context.Context, uuid.UUID, repository.UpdateUserParams) (model.User, error)
	Delete(context.Context, uuid.UUID) error
}

type RegisterInput struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Name     string  `json:"name"`
	Phone    *string `json:"phone"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User model.User `json:"user"`
	AuthTokens
}

type Claims struct {
	UserID uuid.UUID      `json:"user_id"`
	Role   model.UserRole `json:"role"`
	Type   string         `json:"type"`
	jwt.RegisteredClaims
}

func NewAuthService(users userStore, cfg config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	if !validator.ValidateEmail(input.Email) {
		return AuthResponse{}, ValidationError("email", "invalid email")
	}
	if !validator.ValidatePassword(input.Password) {
		return AuthResponse{}, ValidationError("password", "password must contain at least 8 characters")
	}
	if input.Name == "" {
		return AuthResponse{}, ValidationError("name", "name is required")
	}
	if input.Phone != nil && !validator.ValidatePhone(*input.Phone) {
		return AuthResponse{}, ValidationError("phone", "invalid phone")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.users.Create(ctx, repository.CreateUserParams{
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		Name:         input.Name,
		Phone:        input.Phone,
		Role:         model.RoleUser,
	})
	if errors.Is(err, repository.ErrConflict) {
		return AuthResponse{}, ValidationError("email", "email already exists")
	}
	if err != nil {
		return AuthResponse{}, err
	}

	tokens, err := s.issueTokens(user)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{User: user, AuthTokens: tokens}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return AuthResponse{}, ErrUnauthorized
	}
	if err != nil {
		return AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthResponse{}, ErrUnauthorized
	}

	tokens, err := s.issueTokens(user)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{User: user, AuthTokens: tokens}, nil
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (AuthTokens, error) {
	claims, err := s.ParseToken(input.RefreshToken, refreshTokenType)
	if err != nil {
		return AuthTokens{}, ErrUnauthorized
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return AuthTokens{}, ErrUnauthorized
	}
	if err != nil {
		return AuthTokens{}, err
	}

	return s.issueTokens(user)
}

func (s *AuthService) ParseToken(tokenString string, expectedType string) (Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return Claims{}, ErrUnauthorized
	}

	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid || claims.Type != expectedType {
		return Claims{}, ErrUnauthorized
	}
	return claims, nil
}

func (s *AuthService) issueTokens(user model.User) (AuthTokens, error) {
	accessToken, err := s.issueToken(user, accessTokenType, s.cfg.JWTAccessTTL)
	if err != nil {
		return AuthTokens{}, err
	}
	refreshToken, err := s.issueToken(user, refreshTokenType, s.cfg.JWTRefreshTTL)
	if err != nil {
		return AuthTokens{}, err
	}
	return AuthTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthService) issueToken(user model.User, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
