package service

import (
	"context"
	"errors"

	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/http-server/dto"
	"github.com/erknas/ecom/user-service/internal/lib/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserProvider interface {
	UserByID(ctx context.Context, id int64) (*models.User, error)
	UserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
}

type TokenGenerator interface {
	GenerateAccessToken(userID int64, phoneNumber string) (string, error)
	GenerateRefreshToken(userID int64) (string, error)
}

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*jwt.Claims, error)
	ValidateRefreshToken(tokenString string) (int64, error)
}

type TokenRefresher interface {
	RefreshAccessToken(refreshToken string, phoneNumber string) (string, error)
}

type AuthService struct {
	provider  UserProvider
	generator TokenGenerator
	validator TokenValidator
	log       *zap.Logger
}

func NewAuthService(provider UserProvider, generator TokenGenerator, validator TokenValidator, log *zap.Logger) *AuthService {
	return &AuthService{
		provider:  provider,
		generator: generator,
		validator: validator,
		log:       log,
	}
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.provider.UserByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.log.Error("provide user by phone number error", zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password)); err != nil {
		s.log.Warn("invalid credentials", zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.generator.GenerateAccessToken(user.ID, user.PhoneNumber)
	if err != nil {
		s.log.Error("generate access token error", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.generator.GenerateRefreshToken(user.ID)
	if err != nil {
		s.log.Error("generate refresh token error", zap.Error(err))
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	id, err := s.validator.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.log.Error("validate refresh token error", zap.Error(err))
		return nil, err
	}

	user, err := s.provider.UserByID(ctx, id)
	if err != nil {
		s.log.Error("provide user by id error", zap.Error(err))
		return nil, err
	}

	accessToken, err := s.generator.GenerateAccessToken(user.ID, user.PhoneNumber)
	if err != nil {
		s.log.Error("generate access token error", zap.Error(err))
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claim, err := s.validator.ValidateAccessToken(token)
	if err != nil {
		s.log.Error("validate access token error", zap.Error(err))
		return nil, err
	}

	return claim, nil
}
