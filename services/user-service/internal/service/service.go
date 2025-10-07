package service

import (
	"context"
	"errors"

	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/http/dto"
	"github.com/erknas/ecom/user-service/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	Insert(ctx context.Context, user *models.User) (int64, error)
	UserByID(ctx context.Context, id int64) (*models.User, error)
	UserByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id int64, user *models.UpdatedUser) error
}

type TokenGenerator interface {
	GenerateAccessToken(userID int64, email string) (string, error)
}

type Service struct {
	userRepo  UserRepository
	generator TokenGenerator
	log       *zap.Logger
}

func New(userRepo UserRepository, generator TokenGenerator, log *zap.Logger) *Service {
	return &Service{
		userRepo:  userRepo,
		generator: generator,
		log:       log,
	}
}

func (s *Service) CreateNewUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	user, err := models.NewUser(req.FirstName, req.Email, req.Password)
	if err != nil {
		s.log.Error("new user failure", zap.Error(err))
		return nil, err
	}

	id, err := s.userRepo.Insert(ctx, user)
	if err != nil {
		s.log.Error("insert user failure", zap.Error(err))
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
		s.log.Error("user by id failure", zap.Error(err))
		return nil, err
	}

	return &dto.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error) {
	user, err := models.NewUpdatedUser(req.FirstName, req.Email, req.Password)
	if err != nil {
		s.log.Error("new updated user failure", zap.Error(err))
		return nil, err
	}

	if err := s.userRepo.Update(ctx, id, user); err != nil {
		s.log.Error("update user failure", zap.Error(err))
		return nil, err
	}

	return &dto.UpdateUserResponse{
		ID:      id,
		Message: "user updated",
	}, nil
}

func (s *Service) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.UserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Warn("invalid credentials", zap.Error(err))
			return nil, ErrInvalidCredentials
		}

		s.log.Error("user by email failure", zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password)); err != nil {
		s.log.Warn("invalid credentials", zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	token, err := s.generator.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		s.log.Error("generate access token failure", zap.Error(err))
		return nil, err
	}

	return &dto.LoginResponse{
		ID:          user.ID,
		AccessToken: token,
	}, nil
}
