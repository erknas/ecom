package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erknas/ecom/user-service/internal/http/dto"
	mw "github.com/erknas/ecom/user-service/internal/http/middleware"
	"github.com/erknas/ecom/user-service/internal/lib/api"
	"github.com/erknas/ecom/user-service/internal/service"
	"github.com/erknas/ecom/user-service/internal/storage"
	"go.uber.org/zap"
)

type UserService interface {
	CreateNewUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error)
	GetUser(ctx context.Context, id int64) (*dto.User, error)
	UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error)
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

func (h *UserHandlers) HandleRegisterUser(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.CreateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Warn("invalid request body", zap.Error(err))
		return api.InvalidRequestBody()
	}
	defer r.Body.Close()

	if errors := req.Validate(); len(errors) > 0 {
		return api.UnprocessableData(errors)
	}

	resp, err := h.svc.CreateNewUser(r.Context(), req)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			h.log.Warn("create new user failure", zap.Error(err))
			return api.UserAlreadyRegistered()
		}

		return err
	}

	return api.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandlers) HandleGetUserInformation(w http.ResponseWriter, r *http.Request) error {
	id, ok := mw.GetIDFromContext(r.Context())
	if !ok {
		h.log.Warn("user not authorized")
		return api.NotAuthorized()
	}

	resp, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.log.Warn("get user failure", zap.Error(err), zap.Int64("user_id", id))
			return api.UserNotFound()
		}

		return err
	}

	return api.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandlers) HandleUpdateUser(w http.ResponseWriter, r *http.Request) error {
	id, ok := mw.GetIDFromContext(r.Context())
	if !ok {
		h.log.Warn("user not authorized")
		return api.NotAuthorized()
	}

	req := new(dto.UpdateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Warn("invalid request body", zap.Error(err))
		return api.InvalidRequestBody()
	}
	defer r.Body.Close()

	if errors := req.Validate(); len(errors) > 0 {
		return api.UnprocessableData(errors)
	}

	resp, err := h.svc.UpdateUser(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			h.log.Warn("update user failure", zap.Error(err), zap.Int64("user_id", id))
			return api.UserAlreadyRegistered()
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			h.log.Warn("update user failure", zap.Error(err), zap.Int64("user_id", id))
			return api.UserNotFound()
		}
		if errors.Is(err, storage.ErrNoChanges) {
			h.log.Warn("update user failure", zap.Error(err), zap.Int64("user_id", id))
			return api.NothingToUpdate()
		}

		return err
	}

	return api.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandlers) HandleLoginUser(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.LoginRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.log.Warn("invalid request body", zap.Error(err))
		return api.InvalidRequestBody()
	}
	defer r.Body.Close()

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.log.Warn("login failure", zap.Error(err))
			return api.InvalidCredentials()
		}

		return err
	}

	return api.WriteJSON(w, http.StatusOK, resp)
}
