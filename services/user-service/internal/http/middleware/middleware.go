package mw

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
	UserIDKey contextKey = "id"
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

func (m *Middleware) WithJWTAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			m.log.Debug("request header", zap.String("header", header))

			if header == "" {
				m.log.Warn("missing authorization header")
				api.WriteJSON(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			tokenParts := strings.Fields(header)
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				m.log.Warn("invalid header", zap.Any("token parts", tokenParts))
				api.WriteJSON(w, http.StatusUnauthorized, "invalid header format")
				return
			}

			claims, err := m.validator.ValidateAccessToken(r.Context(), tokenParts[1])
			if err != nil {
				m.log.Warn("invalid token")
				api.WriteJSON(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}
