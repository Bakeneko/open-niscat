package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"open-niscat/internal/catalog"
)

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response", "err", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		return // the client went away (e.g. type-ahead search); nothing to report
	}
	switch {
	case errors.Is(err, catalog.ErrNotFound):
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found", Message: err.Error()})
	case errors.Is(err, catalog.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid", Message: err.Error()})
	default:
		slog.Error("request failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal", Message: "internal error"})
	}
}
