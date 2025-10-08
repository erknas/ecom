package service

import (
	"context"
	"testing"
	"time"

	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/http/dto"
	"github.com/erknas/ecom/user-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockUserRepository) Insert(ctx context.Context, user *models.User) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) UserByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id int64, user *models.UpdatedUser) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func (m *MockTokenGenerator) GenerateAccessToken(userID int64, email string) (string, error) {
	args := m.Called(userID, email)
	return args.Get(0).(string), args.Error(1)
}

func TestCreateNewUser(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.CreateUserRequest
		mockSetup   func(mockRepo *MockUserRepository, ctx context.Context)
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, result *dto.CreateUserResponse)
	}{
		{
			name: "success",
			req: &dto.CreateUserRequest{
				FirstName: "User1",
				Email:     "user1@ex.com",
				Password:  "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context) {
				mockRepo.On("Insert", ctx, mock.AnythingOfType("*models.User")).
					Return(int64(1), nil)
			},
			check: func(t *testing.T, result *dto.CreateUserResponse) {
				require.NotNil(t, result)
				assert.Equal(t, int64(1), result.ID)
			},
		},
		{
			name: "check user data is passed correctly",
			req: &dto.CreateUserRequest{
				FirstName: "User2",
				Email:     "user2@ex.com",
				Password:  "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context) {
				mockRepo.On("Insert", ctx, mock.MatchedBy(func(user *models.User) bool {
					if user.FirstName != "User2" {
						return false
					}
					if user.Email != "user2@ex.com" {
						return false
					}
					if len(user.PasswordHash) == 0 {
						return false
					}
					if user.CreatedAt.Truncate(time.Second) != time.Now().Truncate(time.Second) {
						return false
					}
					err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte("password"))
					return err == nil
				})).
					Return(int64(2), nil)
			},
			check: func(t *testing.T, result *dto.CreateUserResponse) {
				assert.Equal(t, int64(2), result.ID)
			},
		},
		{
			name: "duplicate email",
			req: &dto.CreateUserRequest{
				FirstName: "User3",
				Email:     "user2@ex.com",
				Password:  "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context) {
				mockRepo.On("Insert", ctx, mock.AnythingOfType("*models.User")).
					Return(int64(0), storage.ErrUserExists)
			},
			wantErr:     true,
			expectedErr: storage.ErrUserExists,
		},
		{
			name: "internal database error",
			req: &dto.CreateUserRequest{
				FirstName: "User4",
				Email:     "user4@ex.com",
				Password:  "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context) {
				mockRepo.On("Insert", ctx, mock.AnythingOfType("*models.User")).
					Return(int64(0), storage.ErrInternalDatabase)
			},
			wantErr:     true,
			expectedErr: storage.ErrInternalDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			ctx := context.Background()
			log := zap.NewNop()
			svc := New(mockRepo, nil, log)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, ctx)
			}

			result, err := svc.CreateNewUser(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		mockSetup   func(mockRepo *MockUserRepository, ctx context.Context, id int64)
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, result *dto.User, id int64)
	}{
		{
			name:   "success",
			userID: 1,
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("UserByID", ctx, id).
					Return(&models.User{
						ID:        1,
						FirstName: "User1",
						Email:     "user1@ex.com",
						CreatedAt: time.Now(),
					}, nil)
			},
			check: func(t *testing.T, result *dto.User, id int64) {
				require.NotNil(t, result)
				assert.Equal(t, id, result.ID)
				assert.Equal(t, "User1", result.FirstName)
				assert.Equal(t, "user1@ex.com", result.Email)
				assert.Equal(t, time.Now().Truncate(time.Second), result.CreatedAt.Truncate(time.Second))
			},
		},
		{
			name:   "user not found",
			userID: 999,
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("UserByID", ctx, id).
					Return(&models.User{}, storage.ErrUserNotFound)
			},
			wantErr:     true,
			expectedErr: storage.ErrUserNotFound,
		},
		{
			name:   "internal database error",
			userID: 2,
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("UserByID", ctx, id).
					Return(&models.User{}, storage.ErrInternalDatabase)
			},
			wantErr:     true,
			expectedErr: storage.ErrInternalDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			ctx := context.Background()
			log := zap.NewNop()
			svc := New(mockRepo, nil, log)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, ctx, tt.userID)
			}

			result, err := svc.GetUser(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result, tt.userID)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		req         *dto.UpdateUserRequest
		mockSetup   func(mockRepo *MockUserRepository, ctx context.Context, id int64)
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, result *dto.UpdateUserResponse, id int64)
	}{
		{
			name:   "success",
			userID: 1,
			req: &dto.UpdateUserRequest{
				FirstName: strPtr("UpdatedUser1"),
				Email:     strPtr("updateduser1@ex.com"),
				Password:  strPtr("newpassword"),
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("Update", ctx, id, mock.AnythingOfType("*models.UpdatedUser")).
					Return(nil)
			},
			check: func(t *testing.T, result *dto.UpdateUserResponse, id int64) {
				require.NotNil(t, result)
				assert.Equal(t, id, result.ID)
			},
		},
		{
			name:   "check user data is passed correctly",
			userID: 2,
			req: &dto.UpdateUserRequest{
				FirstName: strPtr("UpdatedUser2"),
				Email:     strPtr("updateduser2@ex.com"),
				Password:  strPtr("newpassword"),
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("Update", ctx, id, mock.MatchedBy(func(updatedUser *models.UpdatedUser) bool {
					if updatedUser.FirstName != nil && ptrStr(updatedUser.FirstName) != "UpdatedUser2" {
						return false
					}
					if updatedUser != nil && ptrStr(updatedUser.Email) != "updateduser2@ex.com" {
						return false
					}
					if len(updatedUser.PasswordHash) > 0 {
						err := bcrypt.CompareHashAndPassword(updatedUser.PasswordHash, []byte("newpassword"))
						return err == nil
					}
					return true
				})).
					Return(nil)
			},
			check: func(t *testing.T, result *dto.UpdateUserResponse, id int64) {
				require.NotNil(t, result)
				assert.Equal(t, id, result.ID)
			},
		},
		{
			name:   "duplicate email",
			userID: 3,
			req: &dto.UpdateUserRequest{
				FirstName: strPtr("UpdatedUser3"),
				Email:     strPtr("updateduser1@ex.com"),
				Password:  strPtr("newpassword"),
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("Update", ctx, id, mock.AnythingOfType("*models.UpdatedUser")).
					Return(storage.ErrUserExists)
			},
			wantErr:     true,
			expectedErr: storage.ErrUserExists,
		},
		{
			name:   "no changes",
			userID: 4,
			req:    &dto.UpdateUserRequest{},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("Update", ctx, id, mock.AnythingOfType("*models.UpdatedUser")).
					Return(storage.ErrNoChanges)
			},
			wantErr:     true,
			expectedErr: storage.ErrNoChanges,
		},
		{
			name:   "user not found",
			userID: 999,
			req: &dto.UpdateUserRequest{
				FirstName: strPtr("UpdatedUser999"),
				Email:     strPtr("updateduser@ex.com"),
				Password:  strPtr("newpassword"),
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("Update", ctx, id, mock.AnythingOfType("*models.UpdatedUser")).
					Return(storage.ErrUserNotFound)
			},
			wantErr:     true,
			expectedErr: storage.ErrUserNotFound,
		},
		{
			name:   "internal database error",
			userID: 5,
			req: &dto.UpdateUserRequest{
				FirstName: strPtr("UpdatedUser4"),
			},
			mockSetup: func(mockRepo *MockUserRepository, ctx context.Context, id int64) {
				mockRepo.On("Update", ctx, id, mock.AnythingOfType("*models.UpdatedUser")).
					Return(storage.ErrInternalDatabase)
			},
			wantErr:     true,
			expectedErr: storage.ErrInternalDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			ctx := context.Background()
			log := zap.NewNop()
			svc := New(mockRepo, nil, log)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, ctx, tt.userID)
			}

			result, err := svc.UpdateUser(ctx, tt.userID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result, tt.userID)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.LoginRequest
		mockSetup   func(mockRepo *MockUserRepository, mockGen *MockTokenGenerator, ctx context.Context, email string)
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, result *dto.LoginResponse)
	}{
		{
			name: "success",
			req: &dto.LoginRequest{
				Email:    "user1@ex.com",
				Password: "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, mockGen *MockTokenGenerator, ctx context.Context, email string) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				mockRepo.On("UserByEmail", ctx, email).
					Return(&models.User{
						ID:           1,
						FirstName:    "User1",
						Email:        "user1@ex.com",
						PasswordHash: hash,
						CreatedAt:    time.Now(),
					}, nil)
				mockGen.On("GenerateAccessToken", int64(1), email).
					Return("some.access.token", nil)
			},
			check: func(t *testing.T, result *dto.LoginResponse) {
				require.NotNil(t, result)
				assert.Equal(t, int64(1), result.ID)
				assert.Equal(t, "some.access.token", result.AccessToken)
			},
		},
		{
			name: "invalid credentials password",
			req: &dto.LoginRequest{
				Email:    "user2@ex.com",
				Password: "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, mockGen *MockTokenGenerator, ctx context.Context, email string) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
				mockRepo.On("UserByEmail", ctx, email).
					Return(&models.User{
						ID:           2,
						FirstName:    "User2",
						Email:        "user2@ex.com",
						PasswordHash: hash,
						CreatedAt:    time.Now(),
					}, nil)
				mockGen.AssertNotCalled(t, "GenerateAccessToken")
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "invalid credentials email",
			req: &dto.LoginRequest{
				Email:    "user3@ex.com",
				Password: "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, mockGen *MockTokenGenerator, ctx context.Context, email string) {
				mockRepo.On("UserByEmail", ctx, email).
					Return(&models.User{}, ErrInvalidCredentials)
				mockGen.AssertNotCalled(t, "GenerateAccessToken")
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "internal database error",
			req: &dto.LoginRequest{
				Email:    "user4@ex.com",
				Password: "password",
			},
			mockSetup: func(mockRepo *MockUserRepository, mockGen *MockTokenGenerator, ctx context.Context, email string) {
				mockRepo.On("UserByEmail", ctx, email).
					Return(&models.User{}, storage.ErrInternalDatabase)
				mockGen.AssertNotCalled(t, "GenerateAccessToken")
			},
			wantErr:     true,
			expectedErr: storage.ErrInternalDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockGen := new(MockTokenGenerator)
			ctx := context.Background()
			log := zap.NewNop()
			svc := New(mockRepo, mockGen, log)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockGen, ctx, tt.req.Email)
			}

			result, err := svc.Login(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
			mockGen.AssertExpectations(t)
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func ptrStr(s *string) string {
	return *s
}
