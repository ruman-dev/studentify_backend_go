package utils

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorSource struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ErrorSources []ErrorSource `json:"errorSources,omitempty"`
	Stack        string        `json:"stack,omitempty"`
}

func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	stack := ""
	if err != nil {
		stack = err.Error()
	}

	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Message: message,
		ErrorSources: []ErrorSource{
			{Path: "", Message: message},
		},
		Stack: stack,
	})
}

func ValidationError(w http.ResponseWriter, sources []ErrorSource) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Success:      false,
		Message:      "Validation failed",
		ErrorSources: sources,
	})
}
