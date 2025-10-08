package jwt

import (
	"testing"
	"time"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAccessToken(t *testing.T) {
	tests := []struct {
		name   string
		cfg    func() *config.Config
		userID int64
		email  string
		check  func(t *testing.T, manager *Manager, tokenString string, id int64, email string)
	}{
		{
			name: "success",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Minute * 15,
						Issuer:         "test-issuer",
					},
				}
			},
			userID: 1,
			email:  "user1@email.com",
			check: func(t *testing.T, manager *Manager, tokenString string, id int64, email string) {
				require.NotEmpty(t, tokenString)

				claims, err := manager.ValidateAccessToken(tokenString)
				require.NoError(t, err)
				require.NotNil(t, claims)

				assert.Equal(t, id, claims.UserID)
				assert.Equal(t, email, claims.Email)

				assert.Equal(t, "test-issuer", claims.Issuer)
				assert.False(t, claims.ExpiresAt.Time.IsZero())
				assert.False(t, claims.IssuedAt.Time.IsZero())
				assert.False(t, claims.NotBefore.Time.IsZero())

				now := time.Now()
				assert.True(t, claims.IssuedAt.Before(now.Add(time.Second)))
				assert.True(t, claims.ExpiresAt.After(now))
			},
		},
		{
			name: "check on correct expiration",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Hour,
						Issuer:         "test-issuer",
					},
				}
			},
			userID: 2,
			email:  "user2@email.com",
			check: func(t *testing.T, manager *Manager, tokenString string, id int64, email string) {
				claims, err := manager.ValidateAccessToken(tokenString)
				require.NoError(t, err)

				expectedTime := time.Now().Add(time.Hour)
				actualTime := claims.ExpiresAt.Time

				diff := actualTime.Sub(expectedTime).Abs()
				assert.Less(t, diff, time.Second*5)
			},
		},
		{
			name: "check on unique for different users",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Minute * 15,
						Issuer:         "test-issuer",
					},
				}
			},
			userID: 3,
			email:  "user3@email.com",
			check: func(t *testing.T, manager *Manager, tokenString string, id int64, email string) {
				tokenString2, err := manager.GenerateAccessToken(int64(4), "user4@email.com")
				require.NoError(t, err)

				assert.NotEqual(t, tokenString, tokenString2)

				claims1, err := manager.ValidateAccessToken(tokenString)
				require.NoError(t, err)
				assert.Equal(t, id, claims1.UserID)
				assert.Equal(t, email, claims1.Email)

				claims2, err := manager.ValidateAccessToken(tokenString2)
				require.NoError(t, err)
				assert.Equal(t, int64(4), claims2.UserID)
				assert.Equal(t, "user4@email.com", claims2.Email)
			},
		},
		{
			name: "expires after TTL",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Second,
						Issuer:         "test-issuer",
					},
				}
			},
			userID: 5,
			email:  "user5@email.com",
			check: func(t *testing.T, manager *Manager, tokenString string, id int64, email string) {
				_, err := manager.ValidateAccessToken(tokenString)
				require.NoError(t, err)

				time.Sleep(time.Second * 2)

				_, err = manager.ValidateAccessToken(tokenString)
				require.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrTokenExpired)

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg()
			manager := New(cfg)

			result, err := manager.GenerateAccessToken(tt.userID, tt.email)

			require.NoError(t, err)
			assert.NotEmpty(t, result)
			tt.check(t, manager, result, tt.userID, tt.email)
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		cfg         func() *config.Config
		tokenString func(manager *Manager) string
		wantErr     bool
		expectedErr error
		check       func(t *testing.T, claims *Claims)
	}{
		{
			name: "success",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Minute * 15,
						Issuer:         "test-issuer",
					},
				}
			},
			tokenString: func(manager *Manager) string {
				token, err := manager.GenerateAccessToken(1, "user1@email.com")
				require.NoError(t, err)
				return token
			},
			check: func(t *testing.T, claims *Claims) {
				require.NotNil(t, claims)
				assert.Equal(t, int64(1), claims.UserID)
				assert.Equal(t, "user1@email.com", claims.Email)
				assert.Equal(t, "access_token", claims.Subject)
			},
		},
		{
			name: "invalid format",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Minute * 15,
						Issuer:         "test-issuer",
					},
				}
			},
			tokenString: func(manager *Manager) string {
				return "invalid.access.token"
			},
			wantErr:     true,
			expectedErr: jwt.ErrTokenMalformed,
		},
		{
			name: "expired token",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Millisecond,
						Issuer:         "test-issuer",
					},
				}
			},
			tokenString: func(manager *Manager) string {
				token, err := manager.GenerateAccessToken(2, "user2@email.com")
				require.NoError(t, err)
				return token
			},
			wantErr:     true,
			expectedErr: jwt.ErrTokenExpired,
		},
		{
			name: "empty token",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Millisecond,
						Issuer:         "test-issuer",
					},
				}
			},
			tokenString: func(manager *Manager) string {
				return ""
			},
			wantErr:     true,
			expectedErr: jwt.ErrTokenMalformed,
		},
		{
			name: "invalid subject",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Millisecond,
						Issuer:         "test-issuer",
					},
				}
			},
			tokenString: func(manager *Manager) string {
				claims := Claims{
					UserID: 3,
					Email:  "user3@email.com",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						NotBefore: jwt.NewNumericDate(time.Now()),
						Issuer:    manager.issuer,
						Subject:   "refresh_token",
					},
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString(manager.secret)
				require.NoError(t, err)
				return tokenString
			},
			wantErr:     true,
			expectedErr: jwt.ErrTokenInvalidClaims,
		},
		{
			name: "not valid yet",
			cfg: func() *config.Config {
				return &config.Config{
					JWT: config.JWTConfig{
						Secret:         "test-secret",
						AccessTokenTTL: time.Millisecond,
						Issuer:         "test-issuer",
					},
				}
			},
			tokenString: func(manager *Manager) string {
				claims := Claims{
					UserID: 3,
					Email:  "user3@email.com",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						NotBefore: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						Issuer:    manager.issuer,
						Subject:   "access_token",
					},
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString(manager.secret)
				require.NoError(t, err)
				return tokenString
			},
			wantErr:     true,
			expectedErr: jwt.ErrTokenNotValidYet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg()
			manager := New(cfg)
			token := tt.tokenString(manager)

			result, err := manager.ValidateAccessToken(token)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}
