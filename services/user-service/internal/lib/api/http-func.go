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
			WriteJSON(w, http.StatusBadRequest, map[string]any{
				"status_code": http.StatusBadRequest,
				"error":       err.Error(),
			})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
