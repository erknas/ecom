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

type UserRepository interface {
	InsertUser(ctx context.Context, user *models.User) (int64, error)
	UserByID(ctx context.Context, id int64) (*models.User, error)
	UserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
	Update(ctx context.Context, id int64, user *models.User) error
}

type JWTManager interface {
	GenerateAccessToken(userID int64, phoneNumber string) (string, error)
	GenerateRefreshToken(userID int64) (string, error)
	ValidateAccessToken(tokenString string) (*jwt.Claims, error)
}

type Service struct {
	userRepo   UserRepository
	jwtManager JWTManager
	log        *zap.Logger
}

func New(userRepo UserRepository, jwtManager JWTManager, log *zap.Logger) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		log:        log,
	}
}

func (s *Service) CreateNewUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	user, err := models.NewUser(req.FirstName, req.PhoneNumber, req.Email, req.Password)
	if err != nil {
		s.log.Error("new user error", zap.Error(err))
		return nil, err
	}

	id, err := s.userRepo.InsertUser(ctx, user)
	if err != nil {
		s.log.Error("insert new user error", zap.Error(err))
		return nil, err
	}

	return &dto.CreateUserResponse{
		ID:      id,
		Message: "user created",
	}, nil
}

func (s *Service) GetUser(ctx context.Context, id int64) (*dto.User, error) {
	user, err := s.userRepo.UserByID(ctx, id)
	if err != nil {
		s.log.Error("user by id error", zap.Error(err), zap.Int64("user_id", id))
		return nil, err
	}

	return &dto.User{
		ID:          user.ID,
		FirstName:   user.FirstName,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.UserByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		s.log.Error("user by phone number error", zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password)); err != nil {
		s.log.Warn("invalid credentials", zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, req.PhoneNumber)
	if err != nil {
		s.log.Error("generate access token error", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		s.log.Error("generate refresh token error", zap.Error(err))
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		s.log.Error("validate access token error", zap.Error(err))
		return nil, err
	}

	return claims, nil
}

func (s *Service) UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) error {
	user, err := models.NewUser(req.FirstName, req.PhoneNumber, req.Email, req.Password)
	if err != nil {
		s.log.Error("new user error", zap.Error(err))
		return err
	}

	if err := s.userRepo.Update(ctx, id, user); err != nil {
		s.log.Error("update user error", zap.Error(err), zap.Int64("user_id", id))
		return err
	}

	return nil
}
