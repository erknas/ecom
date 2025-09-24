package service

import (
	"context"

	"github.com/erknas/ecom/user-service/internal/domain/models"
)

type UserInserter interface {
	InsertUser(ctx context.Context, user *models.User) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, id int64) (*models.User, error)
}

type UserService struct {
	userInserter UserInserter
	userProvider UserProvider
}

func New(userInserter UserInserter, userProvider UserProvider) *UserService {
	return &UserService{
		userInserter: userInserter,
		userProvider: userProvider,
	}
}

func (s *UserService) CreateNewUser(ctx context.Context, firstName, phoneNumber, email, password string) (int64, error) {
	user, err := models.NewUser(firstName, phoneNumber, email, password)
	if err != nil {
		return 0, err
	}

	id, err := s.userInserter.InsertUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.userProvider.User(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
