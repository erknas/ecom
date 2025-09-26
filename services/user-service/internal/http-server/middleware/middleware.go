package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/erknas/ecom/user-service/internal/lib/api"
	"github.com/erknas/ecom/user-service/internal/lib/jwt"
	"go.uber.org/zap"
)

type contextKey string

const (
	UserIDKey          contextKey = "id"
	UserPhoneNumberKey contextKey = "phone_number"
)

type TokenValidator interface {
	ValidateAccessToken(ctx context.Context, token string) (*jwt.Claims, error)
}

type Middleware struct {
	validator TokenValidator
	log       *zap.Logger
}

func New(validator TokenValidator, log *zap.Logger) *Middleware {
	return &Middleware{
		validator: validator,
		log:       log,
	}
}

func (m *Middleware) WithJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.log.Warn("missing authorization header")
			api.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid header"})
			return
		}

		tokenFileds := strings.SplitN(authHeader, " ", 2)
		if len(tokenFileds) != 2 || tokenFileds[0] != "Bearer" {
			api.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid header"})
			m.log.Warn("invalid authorization header format")
			return
		}

		claims, err := m.validator.ValidateAccessToken(r.Context(), tokenFileds[1])
		if err != nil {
			m.log.Warn("invalid token", zap.Error(err))
			api.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid token"})
			return
		}

		m.log.Debug("access token", zap.String("token", tokenFileds[1]))

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserPhoneNumberKey, claims.PhoneNumber)

		next(w, r.WithContext(ctx))
	}
}

func GetIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}

func GetPhoneNumberFromContext(ctx context.Context) (string, bool) {
	phoneNumber, ok := ctx.Value(UserPhoneNumberKey).(string)
	return phoneNumber, ok
}
