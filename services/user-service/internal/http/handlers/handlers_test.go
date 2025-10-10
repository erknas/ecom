package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erknas/ecom/user-service/internal/http/dto"
	mw "github.com/erknas/ecom/user-service/internal/http/middleware"
	"github.com/erknas/ecom/user-service/internal/lib/api"
	"github.com/erknas/ecom/user-service/internal/service"
	"github.com/erknas/ecom/user-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateNewUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dto.CreateUserResponse), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id int64) (*dto.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*dto.UpdateUserResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func TestHandleRegisterUser(t *testing.T) {
	tests := []struct {
		name           string
		req            string
		mockSetup      func(mockSvc *MockUserService)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			req:  `{"first_name": "User1", "email": "user1@ex.com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("CreateNewUser", mock.Anything, mock.MatchedBy(func(req *dto.CreateUserRequest) bool {
					return req.FirstName == "User1" && req.Email == "user1@ex.com" && req.Password == "password"
				})).Return(&dto.CreateUserResponse{
					ID:      1,
					Message: "user created",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			check: func(t *testing.T, body []byte) {
				var resp dto.CreateUserResponse
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, "user created", resp.Message)
			},
		},
		{
			name: "invalid JSON",
			req:  `{"first_name: User2,}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "CreateNewUser")
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "invalid request body")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "empty request body",
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "CreateNewUser")
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "invalid request body")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "validation missing first name",
			req:  `{"email": "user1@ex.com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "CreateNewUser")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "first_name")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "validation invlaid email format",
			req:  `{"first_name": "User", "email": "user-ex-com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "CreateNewUser")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "email")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "validation short password",
			req:  `{"first_name": "User", "email": "user@ex.com", "password": "pass"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "CreateNewUser")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "password")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "user already exists",
			req:  `{"first_name": "User1", "email": "user1@ex.com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("CreateNewUser", mock.Anything, mock.MatchedBy(func(req *dto.CreateUserRequest) bool {
					return req.FirstName == "User1" && req.Email == "user1@ex.com" && req.Password == "password"
				})).Return(&dto.CreateUserResponse{}, storage.ErrUserExists)
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "user already registered")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "internal error",
			req:  `{"first_name": "User1", "email": "user1@ex.com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("CreateNewUser", mock.Anything, mock.AnythingOfType("*dto.CreateUserRequest")).
					Return(&dto.CreateUserResponse{}, errors.New("some internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "unexpected error")
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockUserService)
			log := zap.NewNop()
			handlers := New(mockSvc, log)

			tt.mockSetup(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(tt.req)))
			rr := httptest.NewRecorder()
			defer rr.Result().Body.Close()

			api.MakeHTTPFunc(handlers.HandleRegisterUser).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			body := rr.Body.Bytes()
			tt.check(t, body)

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetUserInformation(t *testing.T) {
	tests := []struct {
		name           string
		ctxSetup       func() context.Context
		mockSetup      func(mockSvc *MockUserService)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 1)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("GetUser", mock.Anything, int64(1)).
					Return(&dto.User{
						ID:        1,
						FirstName: "User1",
						Email:     "user1@ex.com",
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.User
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, "User1", resp.FirstName)
				assert.Equal(t, "user1@ex.com", resp.Email)
				assert.Equal(t, time.Now().Truncate(time.Second), resp.CreatedAt.Truncate(time.Second))
			},
		},
		{
			name: "user not authorized",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "GetUser")
			},
			expectedStatus: http.StatusUnauthorized,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "not authorized")
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "user not found",
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 2)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("GetUser", mock.Anything, int64(2)).
					Return(&dto.User{}, storage.ErrUserNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "user not found")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "internal error",
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 3)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("GetUser", mock.Anything, int64(3)).
					Return(&dto.User{}, errors.New("some internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "unexpected error")
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
		mockSvc := new(MockUserService)
		log := zap.NewNop()
		handlers := New(mockSvc, log)

		tt.mockSetup(mockSvc)

		ctx := tt.ctxSetup()

		req := httptest.NewRequest(http.MethodGet, "/api/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()
		defer rr.Result().Body.Close()

		api.MakeHTTPFunc(handlers.HandleGetUserInformation).ServeHTTP(rr, req)

		assert.Equal(t, tt.expectedStatus, rr.Code)

		body := rr.Body.Bytes()
		tt.check(t, body)

		mockSvc.AssertExpectations(t)
	}
}

func TestHandleUpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		req            string
		ctxSetup       func() context.Context
		mockSetup      func(mockSvc *MockUserService)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			req:  `{"first_name": "Newuser1", "email": "newuser1@ex.com", "password": "newpassword"}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 1)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("UpdateUser", mock.Anything, int64(1), mock.MatchedBy(func(req *dto.UpdateUserRequest) bool {
					return ptrStr(req.FirstName) == "Newuser1" && ptrStr(req.Email) == "newuser1@ex.com" && ptrStr(req.Password) == "newpassword"
				})).Return(&dto.UpdateUserResponse{
					ID:      1,
					Message: "user updated",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.UpdateUserResponse
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, "user updated", resp.Message)
			},
		},
		{
			name: "user not authorized",
			req:  `{"first_name": "Newuser1", "email": "newuser1@ex.com", "password": "newpassword"}`,
			ctxSetup: func() context.Context {
				return context.Background()
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "UpdateUser")
			},
			expectedStatus: http.StatusUnauthorized,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "not authorized")
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "invalid JSON",
			req:  `{"first_name: User2,}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 3)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "UpdateUser")
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "invalid request body")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "empty request body",
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 4)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "UpdateUser")
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "invalid request body")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "validation first name error",
			req:  `{"first_name": "", "email": "user1@ex.com", "password": "password"}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 4)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "UpdateUser")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "first_name")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "validation invlaid email format",
			req:  `{"first_name": "User", "email": "user-ex-com", "password": "password"}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 4)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "UpdateUser")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "email")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "validation short password",
			req:  `{"first_name": "User", "email": "user@ex.com", "password": "pass"}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 4)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "UpdateUser")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "password")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "user already exists",
			req:  `{"first_name": "User1", "email": "user1@ex.com", "password": "password"}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 4)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("UpdateUser", mock.Anything, int64(4), mock.MatchedBy(func(req *dto.UpdateUserRequest) bool {
					return ptrStr(req.FirstName) == "User1" && ptrStr(req.Email) == "user1@ex.com" && ptrStr(req.Password) == "password"
				})).Return(&dto.UpdateUserResponse{}, storage.ErrUserExists)
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "user already registered")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "nothing to update",
			req:  `{}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 4)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "UpdateUser")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "empty_fields")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "internal error",
			req:  `{"first_name": "Newuser1", "email": "newuser1@ex.com", "password": "newpassword"}`,
			ctxSetup: func() context.Context {
				return mw.SetIDInContext(context.Background(), 1)
			},
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("UpdateUser", mock.Anything, mock.Anything, mock.AnythingOfType("*dto.UpdateUserRequest")).
					Return(&dto.UpdateUserResponse{}, errors.New("some internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "unexpected error")
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockUserService)
			log := zap.NewNop()
			handlers := New(mockSvc, log)

			tt.mockSetup(mockSvc)

			ctx := tt.ctxSetup()

			req := httptest.NewRequest(http.MethodPut, "/api/me/update", bytes.NewReader([]byte(tt.req))).WithContext(ctx)
			rr := httptest.NewRecorder()
			defer rr.Result().Body.Close()

			api.MakeHTTPFunc(handlers.HandleUpdateUser).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			body := rr.Body.Bytes()
			tt.check(t, body)

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandleLoginUser(t *testing.T) {
	tests := []struct {
		name           string
		req            string
		mockSetup      func(mockSvc *MockUserService)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			req:  `{"email": "user1@ex.com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("Login", mock.Anything, mock.MatchedBy(func(req *dto.LoginRequest) bool {
					return req.Email == "user1@ex.com" && req.Password == "password"
				})).Return(&dto.LoginResponse{
					ID:          1,
					AccessToken: "user1.access.token",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.LoginResponse
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Equal(t, int64(1), resp.ID)
				assert.Equal(t, "user1.access.token", resp.AccessToken)
			},
		},
		{
			name: "invalid request body",
			req:  `{"email "user1@ex.com" password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "Login")
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "invalid request body")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "empty request body",
			req:  `{"email "user1@ex.com" password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "Login")
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "invalid request body")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "invalid credentials",
			req:  `{"email": "user1@ex.com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("Login", mock.Anything, mock.MatchedBy(func(req *dto.LoginRequest) bool {
					return req.Email == "user1@ex.com" && req.Password == "password"
				})).Return(&dto.LoginResponse{}, service.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusBadRequest,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "invalid email or password")
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "empty email",
			req:  `{"email": "", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "Login")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "email")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "empty password",
			req:  `{"email": "user1@ex.com", "password": ""}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "Login")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "password")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "empty email and password",
			req:  `{"email": "", "password": ""}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.AssertNotCalled(t, "Login")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "email")
				assert.Contains(t, resp.Message, "password")
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			},
		},
		{
			name: "internal error",
			req:  `{"email": "user1@ex.com", "password": "password"}`,
			mockSetup: func(mockSvc *MockUserService) {
				mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).
					Return(&dto.LoginResponse{}, errors.New("some internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			check: func(t *testing.T, body []byte) {
				var resp api.APIError
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Contains(t, resp.Message, "unexpected error")
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockUserService)
			log := zap.NewNop()
			handlers := New(mockSvc, log)

			tt.mockSetup(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte(tt.req)))
			rr := httptest.NewRecorder()
			defer rr.Result().Body.Close()

			api.MakeHTTPFunc(handlers.HandleLoginUser).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			body := rr.Body.Bytes()
			tt.check(t, body)
		})
	}
}

func ptrStr(s *string) string {
	return *s
}
