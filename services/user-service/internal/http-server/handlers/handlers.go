package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erknas/ecom/user-service/internal/http-server/dto"
	"github.com/erknas/ecom/user-service/internal/http-server/middleware"
	"github.com/erknas/ecom/user-service/internal/lib/api"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserService interface {
	CreateNewUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error)
	GetUser(ctx context.Context, id int64) (*dto.User, error)
}

type AuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
}

type Middleware interface {
	WithJWTAuth(next http.HandlerFunc) http.HandlerFunc
}

type UserHandler struct {
	user UserService
	auth AuthService
	mw   Middleware
	log  *zap.Logger
}

func New(user UserService, auth AuthService, mw Middleware, log *zap.Logger) *UserHandler {
	return &UserHandler{
		user: user,
		auth: auth,
		mw:   mw,
		log:  log,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/create", api.MakeHTTPFunc(h.handleRegisterUser))
	r.Post("/login", api.MakeHTTPFunc(h.handleLogin))

	r.Get("/me", h.mw.WithJWTAuth(api.MakeHTTPFunc(h.handleGetUser)))
}

func (h *UserHandler) handleRegisterUser(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.CreateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Error("decode request body error", zap.Error(err))
		return err
	}
	defer r.Body.Close()

	resp, err := h.user.CreateNewUser(r.Context(), req)
	if err != nil {
		h.log.Error("register user error", zap.Error(err))
		return err
	}

	return api.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) handleGetUser(w http.ResponseWriter, r *http.Request) error {
	id, ok := middleware.GetIDFromContext(r.Context())
	if !ok {
		return api.WriteJSON(w, http.StatusUnauthorized, "not authenticated")
	}

	user, err := h.user.GetUser(r.Context(), id)
	if err != nil {
		h.log.Error("get user error", zap.Error(err))
		return err
	}

	return api.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.LoginRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Error("decode request body error", zap.Error(err))
		return err
	}
	defer r.Body.Close()

	resp, err := h.auth.Login(r.Context(), req)
	if err != nil {
		h.log.Error("login error", zap.Error(err))
		return err
	}

	return api.WriteJSON(w, http.StatusOK, resp)
}
