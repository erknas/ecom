package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/http-server/dto"
	"github.com/erknas/ecom/user-service/internal/lib/api"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Service interface {
	CreateNewUser(ctx context.Context, firstName, phoneNumber, email, password string) (int64, error)
	GetUser(ctx context.Context, id int64) (*models.User, error)
}

type UserHandler struct {
	svc Service
	log *zap.Logger
}

func New(svc Service, log *zap.Logger) *UserHandler {
	return &UserHandler{
		svc: svc,
		log: log,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/create", api.MakeHTTPFunc(h.handleRegisterUser))
}

func (h *UserHandler) handleRegisterUser(w http.ResponseWriter, r *http.Request) error {
	userReq := new(dto.CreateUserRequest)

	if err := json.NewDecoder(r.Body).Decode(userReq); err != nil {
		h.log.Error("decode request body error",
			zap.Error(err),
		)
		return err
	}
	defer r.Body.Close()

	id, err := h.svc.CreateNewUser(r.Context(), userReq.FirstName, userReq.PhoneNumber, userReq.Email, userReq.Password)
	if err != nil {
		h.log.Error("create new user error",
			zap.Error(err),
		)
		return err
	}

	userResp := dto.CreateUserResponse{
		ID:      id,
		Message: "user created",
	}

	return api.WriteJSON(w, http.StatusCreated, userResp)
}
