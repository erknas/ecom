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

func TestInsertUser(t *testing.T) {
	db := NewMockStorage()

	tests := []struct {
		name        string
		user        *models.User
		ctx         func() context.Context
		wantErr     bool
		containsErr error
	}{
		{
			name: "successful insert user",
			user: &models.User{
				FirstName:    "User1",
				Email:        "user@example.com",
				PasswordHash: []byte("password"),
			},
			wantErr: false,
		},
		{
			name: "error duplicate email",
			user: &models.User{
				FirstName:    "User2",
				Email:        "user@example.com",
				PasswordHash: []byte("password"),
			},
			wantErr:     true,
			containsErr: storage.ErrUserExists,
		},
		{
			name: "error context timeout",
			user: &models.User{
				FirstName:    "User3",
				Email:        "user3@example.com",
				PasswordHash: []byte("password"),
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()
				time.Sleep(time.Millisecond)
				return ctx
			},
			wantErr:     true,
			containsErr: context.DeadlineExceeded,
		},
		{
			name: "error context cancelled",
			user: &models.User{
				FirstName:    "User4",
				Email:        "user4@example.com",
				PasswordHash: []byte("password"),
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			wantErr:     true,
			containsErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.ctx != nil {
				ctx = tt.ctx()
			}

			id, err := db.InsertUser(ctx, tt.user)
			if tt.wantErr {
				require.Error(t, err)
				assert.Zero(t, id)
				assert.ErrorIs(t, err, tt.containsErr)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, id)
			}
		})
	}
}

func TestUserByID(t *testing.T) {

}
