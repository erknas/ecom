package api

import (
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int `json:"status_code"`
	Message    any `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

func NewAPIError(status int, err error) APIError {
	return APIError{
		StatusCode: status,
		Message:    err.Error(),
	}
}

func InvalidRequestBody() APIError {
	return NewAPIError(http.StatusBadRequest, fmt.Errorf("invalid request body"))
}

func UnprocessableData(errors map[string]string) APIError {
	return APIError{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    errors,
	}
}

func InternalError() APIError {
	return NewAPIError(http.StatusInternalServerError, fmt.Errorf("unexpected error"))
}

func NotAuthorized() APIError {
	return NewAPIError(http.StatusUnauthorized, fmt.Errorf("not authorized"))
}

func InvalidCredentials() APIError {
	return NewAPIError(http.StatusBadRequest, fmt.Errorf("invalid email or password"))
}

func UserAlreadyRegistered() APIError {
	return NewAPIError(http.StatusBadRequest, fmt.Errorf("user already registered"))
}

func UserNotFound() APIError {
	return NewAPIError(http.StatusBadRequest, fmt.Errorf("user not found"))
}

func NothingToUpdate() APIError {
	return NewAPIError(http.StatusBadRequest, fmt.Errorf("nothing to update"))
}
