package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUser(t *testing.T) {
	t.Run("hashes password correctly", func(t *testing.T) {
		password := "password123"
		user, err := NewUser("User", "user@example.com", password)

		require.NoError(t, err)

		err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
		assert.NoError(t, err)

		assert.NotEqual(t, password, string(user.PasswordHash))
	})
}
