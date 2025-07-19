package client

import "errors"

// Sentinel ошибки
var (
	ErrMissingCredentials = errors.New("username and password must be set")
	ErrTokenNotFound      = errors.New("access_token not found in response")
	ErrTokenRefreshFailed = errors.New("failed to refresh token")
	ErrRequestFailed      = errors.New("request failed")
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrInvalidAPIResponse = errors.New("invalid API response")
	ErrMethodNotAllowed   = errors.New("method not allowed (405)")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidURL         = errors.New("invalid URL: contains spaces or malformed")
	ErrEndpointNotFound   = errors.New("API endpoint not found (404)")
)
