package service

import (
	"context"

	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/http-server/dto"
	"go.uber.org/zap"
)

type UserInserter interface {
	InsertUser(ctx context.Context, user *models.User) (int64, error)
}

type UserService struct {
	inserter UserInserter
	provider UserProvider
	log      *zap.Logger
}

func NewUserService(inserter UserInserter, provider UserProvider, log *zap.Logger) *UserService {
	return &UserService{
		inserter: inserter,
		provider: provider,
		log:      log,
	}
}

func (s *UserService) CreateNewUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	user, err := models.NewUser(req.FirstName, req.PhoneNumber, req.Email, req.Password)
	if err != nil {
		s.log.Error("new user error", zap.Error(err))
		return nil, err
	}

	id, err := s.inserter.InsertUser(ctx, user)
	if err != nil {
		s.log.Error("insert user error", zap.Error(err))
		return nil, err
	}

	return &dto.CreateUserResponse{
		ID:      id,
		Message: "user created",
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*dto.User, error) {
	user, err := s.provider.UserByID(ctx, id)
	if err != nil {
		s.log.Error("prvoide user by id error", zap.Error(err))
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
