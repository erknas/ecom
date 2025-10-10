package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type httpFunc func(w http.ResponseWriter, r *http.Request) error

func MakeHTTPFunc(fn httpFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		if err := fn(w, r.WithContext(ctx)); err != nil {
			if apiErr, ok := err.(APIError); ok {
				WriteJSON(w, apiErr.StatusCode, apiErr)
			} else {
				errResp := map[string]any{
					"status_code": http.StatusInternalServerError,
					"message":     "unexpected error",
				}
				WriteJSON(w, http.StatusInternalServerError, errResp)
			}
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
