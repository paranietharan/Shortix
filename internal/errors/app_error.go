package errors

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

func InternalServerError() *AppError {
	return New("INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
}
