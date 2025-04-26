package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ServerError struct {
	Err  error
	Code int
}

func NewError(err error, code int) *ServerError {
	return &ServerError{
		Err:  err,
		Code: code,
	}
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%s [%d]", e.Err.Error(), e.Code)
}

func (e *ServerError) Write(w http.ResponseWriter) {
	resp := ErrorResponse{
		Code:    e.Code,
		Message: e.Err.Error(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response json", slog.Any("error", err), slog.Any("resp", resp))
	}
}
