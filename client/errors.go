package client

import (
	"errors"
	"net/http"
)

var (
	ErrUnauthorized = errors.New("band: unauthorized")
	ErrForbidden    = errors.New("band: forbidden")
	ErrNotFound     = errors.New("band: not found")
	ErrRateLimited  = errors.New("band: rate limited")
	ErrServer       = errors.New("band: server error")
)

// ApiError describes a non-2xx API response.
type ApiError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	RetryAfter string
	Body       string
}

func (err *ApiError) Error() string {
	if err.Message != "" {
		return err.Message
	}
	if err.Code != "" {
		return err.Code
	}
	return http.StatusText(err.StatusCode)
}

func (err *ApiError) Unwrap() error {
	switch err.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		if err.StatusCode >= http.StatusInternalServerError {
			return ErrServer
		}
		return nil
	}
}
