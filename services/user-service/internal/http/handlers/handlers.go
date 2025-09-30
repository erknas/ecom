package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erknas/ecom/user-service/internal/http/dto"
	mw "github.com/erknas/ecom/user-service/internal/http/middleware"
	"github.com/erknas/ecom/user-service/internal/lib/api"
	"go.uber.org/zap"
)

type UserService interface {
	CreateNewUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error)
	GetUser(ctx context.Context, id int64) (*dto.User, error)
	UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) error
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
}

type UserHandlers struct {
	svc UserService
	log *zap.Logger
}

func New(svc UserService, log *zap.Logger) *UserHandlers {
	return &UserHandlers{
		svc: svc,
		log: log,
	}
}

func (h *UserHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.CreateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Warn("invalid request body", zap.Error(err))
		return err
	}
	defer r.Body.Close()

	resp, err := h.svc.CreateNewUser(r.Context(), req)
	if err != nil {
		h.log.Error("register new user error", zap.Error(err))
		return err
	}

	return api.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandlers) HandleGetUser(w http.ResponseWriter, r *http.Request) error {
	id, ok := mw.GetIDFromContext(r.Context())
	if !ok {
		h.log.Warn("user not found", zap.Int64("user_id", id))
		return api.WriteJSON(w, http.StatusUnauthorized, "not authorized")
	}

	resp, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		h.log.Error("get user error", zap.Error(err))
		return err
	}

	return api.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.LoginRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Warn("invalid request body", zap.Error(err))
		return err
	}
	defer r.Body.Close()

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		h.log.Error("login error", zap.Error(err))
	}

	return api.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandlers) HandleUpdateUser(w http.ResponseWriter, r *http.Request) error {
	id, ok := mw.GetIDFromContext(r.Context())
	if !ok {
		h.log.Warn("user not found", zap.Int64("user_id", id))
		return api.WriteJSON(w, http.StatusUnauthorized, "not authorized")
	}

	req := new(dto.UpdateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Warn("invalid request body", zap.Error(err))
		return err
	}
	defer r.Body.Close()

	if err := h.svc.UpdateUser(r.Context(), id, req); err != nil {
		h.log.Error("update user error", zap.Error(err))
		return err
	}

	return api.WriteJSON(w, http.StatusOK, "user updated")
}
