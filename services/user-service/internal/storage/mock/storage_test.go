package mock

import (
	"context"
	"testing"
	"time"

	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(db *MockStorage)
		user        *models.User
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, db *MockStorage, id int64)
	}{
		{
			name:  "success",
			setup: nil,
			user: &models.User{
				FirstName:    "User1",
				Email:        "user1@ex.com",
				PasswordHash: []byte("password"),
				CreatedAt:    time.Now(),
			},
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, db *MockStorage, id int64) {
				assert.Equal(t, int64(1), id)

				user, err := db.UserByID(context.Background(), id)
				require.NoError(t, err)

				assert.Equal(t, id, int64(1))
				assert.Equal(t, "User1", user.FirstName)
				assert.Equal(t, "user1@ex.com", user.Email)
				assert.Equal(t, []byte("password"), user.PasswordHash)
				assert.Equal(t, time.Now().Truncate(time.Second), user.CreatedAt.Truncate(time.Second))
			},
		},
		{
			name: "duplicate email",
			setup: func(db *MockStorage) {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				_, err := db.Insert(context.Background(), user)
				require.NoError(t, err)
			},
			user: &models.User{
				FirstName:    "User2",
				Email:        "user1@ex.com",
				PasswordHash: []byte("password"),
				CreatedAt:    time.Now(),
			},
			wantErr:     true,
			expectedErr: storage.ErrUserExists,
			check:       nil,
		},
		{
			name: "generate unique user id",
			setup: func(db *MockStorage) {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id, err := db.Insert(context.Background(), user)
				require.NoError(t, err)
				assert.Equal(t, int64(1), id)
			},
			user: &models.User{
				FirstName:    "User2",
				Email:        "user2@ex.com",
				PasswordHash: []byte("password"),
				CreatedAt:    time.Now(),
			},
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, db *MockStorage, id int64) {
				assert.Equal(t, int64(2), id)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockStorage()

			if tt.setup != nil {
				tt.setup(db)
			}

			id, err := db.Insert(context.Background(), tt.user)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Zero(t, id)
			} else {
				require.NoError(t, err)
				assert.Greater(t, id, int64(0))
				if tt.check != nil {
					tt.check(t, db, id)
				}
			}
		})
	}
}

func TestUserByID(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(db *MockStorage) int64
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, user *models.User)
	}{
		{
			name: "success",
			setup: func(db *MockStorage) int64 {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id, err := db.Insert(context.Background(), user)
				require.NoError(t, err)
				return id
			},
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, user *models.User) {
				assert.Equal(t, int64(1), user.ID)
				assert.Equal(t, "User1", user.FirstName)
				assert.Equal(t, "user1@ex.com", user.Email)
				assert.Equal(t, []byte("password"), user.PasswordHash)
				assert.Equal(t, time.Now().Truncate(time.Second), user.CreatedAt.Truncate(time.Second))
			},
		},
		{
			name: "user not found",
			setup: func(db *MockStorage) int64 {
				return 999
			},
			wantErr:     true,
			expectedErr: storage.ErrUserNotFound,
			check:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockStorage()
			userID := tt.setup(db)

			user, err := db.UserByID(context.Background(), userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, user)
				}
			}
		})
	}
}

func TestUserByEmail(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(db *MockStorage)
		email       string
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, user *models.User)
	}{
		{
			name: "success",
			setup: func(db *MockStorage) {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				_, err := db.Insert(context.Background(), user)
				require.NoError(t, err)
			},
			email:       "user1@ex.com",
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, user *models.User) {
				assert.Equal(t, int64(1), user.ID)
				assert.Equal(t, "User1", user.FirstName)
				assert.Equal(t, "user1@ex.com", user.Email)
				assert.Equal(t, []byte("password"), user.PasswordHash)
				assert.Equal(t, time.Now().Truncate(time.Second), user.CreatedAt.Truncate(time.Second))
			},
		},
		{
			name: "user not found",
			setup: func(db *MockStorage) {
				user := &models.User{
					FirstName:    "User2",
					Email:        "user2@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				_, err := db.Insert(context.Background(), user)
				require.NoError(t, err)
			},
			email:       "user1@ex.com",
			wantErr:     true,
			expectedErr: storage.ErrUserNotFound,
			check:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockStorage()
			tt.setup(db)

			user, err := db.UserByEmail(context.Background(), tt.email)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, user)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(db *MockStorage) int64
		update      *models.UpdatedUser
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, db *MockStorage, id int64)
	}{
		{
			name: "update all fields",
			setup: func(db *MockStorage) int64 {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id, _ := db.Insert(context.Background(), user)
				return id
			},
			update: &models.UpdatedUser{
				FirstName:    strPtr("User2"),
				Email:        strPtr("user2@ex.com"),
				PasswordHash: []byte("newpassword"),
			},
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, db *MockStorage, id int64) {
				user, err := db.UserByID(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "User2", user.FirstName)
				assert.Equal(t, "user2@ex.com", user.Email)
				assert.Equal(t, []byte("newpassword"), user.PasswordHash)
			},
		},
		{
			name: "update first name",
			setup: func(db *MockStorage) int64 {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id, _ := db.Insert(context.Background(), user)
				return id
			},
			update: &models.UpdatedUser{
				FirstName: strPtr("User2"),
			},
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, db *MockStorage, id int64) {
				user, err := db.UserByID(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "User2", user.FirstName)
				assert.Equal(t, "user1@ex.com", user.Email)
				assert.Equal(t, []byte("password"), user.PasswordHash)
			},
		},
		{
			name: "update email",
			setup: func(db *MockStorage) int64 {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id, _ := db.Insert(context.Background(), user)
				return id
			},
			update: &models.UpdatedUser{
				Email: strPtr("user2@ex.com"),
			},
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, db *MockStorage, id int64) {
				user, err := db.UserByID(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "User1", user.FirstName)
				assert.Equal(t, "user2@ex.com", user.Email)
				assert.Equal(t, []byte("password"), user.PasswordHash)
			},
		},
		{
			name: "update password",
			setup: func(db *MockStorage) int64 {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id, _ := db.Insert(context.Background(), user)
				return id
			},
			update: &models.UpdatedUser{
				PasswordHash: []byte("newpassword"),
			},
			wantErr:     false,
			expectedErr: nil,
			check: func(t *testing.T, db *MockStorage, id int64) {
				user, err := db.UserByID(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "User1", user.FirstName)
				assert.Equal(t, "user1@ex.com", user.Email)
				assert.Equal(t, []byte("newpassword"), user.PasswordHash)
			},
		},
		{
			name: "no changes",
			setup: func(db *MockStorage) int64 {
				user := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id, _ := db.Insert(context.Background(), user)
				return id
			},
			update:      &models.UpdatedUser{},
			wantErr:     true,
			expectedErr: storage.ErrNoChanges,
			check:       nil,
		},
		{
			name: "duplicate email",
			setup: func(db *MockStorage) int64 {
				user1 := &models.User{
					FirstName:    "User1",
					Email:        "user1@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				user2 := &models.User{
					FirstName:    "User2",
					Email:        "user2@ex.com",
					PasswordHash: []byte("password"),
					CreatedAt:    time.Now(),
				}
				id1, _ := db.Insert(context.Background(), user1)
				_, _ = db.Insert(context.Background(), user2)
				return id1
			},
			update: &models.UpdatedUser{
				Email: strPtr("user2@ex.com"),
			},
			wantErr:     true,
			expectedErr: storage.ErrUserExists,
			check:       nil,
		},
		{
			name: "user not found",
			setup: func(db *MockStorage) int64 {
				return 999
			},
			update: &models.UpdatedUser{
				FirstName:    strPtr("User2"),
				Email:        strPtr("user2@ex.com"),
				PasswordHash: []byte("newpassword"),
			},
			wantErr:     true,
			expectedErr: storage.ErrUserNotFound,
			check:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockStorage()
			userID := tt.setup(db)

			err := db.Update(context.Background(), userID, tt.update)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, db, userID)
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
