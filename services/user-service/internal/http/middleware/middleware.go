package mw

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/erknas/ecom/user-service/internal/lib/api"
	"github.com/erknas/ecom/user-service/internal/lib/jwt"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type contextKey string

const (
	UserIDKey contextKey = "id"
)

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*jwt.Claims, error)
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
			if header == "" {
				m.log.Warn("missing authorization header")
				permissionDenied(w)
				return
			}

			tokenParts := strings.Fields(header)
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				m.log.Warn("invalid header", zap.String("header", header))
				permissionDenied(w)
				return
			}

			claims, err := m.validator.ValidateAccessToken(tokenParts[1])
			if err != nil {
				m.log.Warn("invalid token")
				permissionDenied(w)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *Middleware) WithLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			entry := m.log.With(
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.String("request_id", middleware.GetReqID(r.Context())),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			defer func() {
				entry.Info("request completed",
					zap.Int("status", ww.Status()),
					zap.Int("bytes", ww.BytesWritten()),
					zap.Duration("duration", time.Since(start)),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func GetIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}

func permissionDenied(w http.ResponseWriter) {
	api.WriteJSON(w, http.StatusUnauthorized, api.APIError{
		StatusCode: http.StatusUnauthorized,
		Message:    "authentication required"},
	)
}
