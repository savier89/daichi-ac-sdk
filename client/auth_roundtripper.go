package client

import (
	"context"
	"net/http"
)

// AuthRoundTripper добавляет токен к каждому запросу
type AuthRoundTripper struct {
	Transport http.RoundTripper
	Token     string
	RefreshFn func(context.Context) (string, error)
	Logger    *Logger
}

// RoundTrip реализует интерфейс http.RoundTripper
func (rt *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.Token != "" {
		req = req.Clone(req.Context())
		req.Header.Set("Authorization", "Bearer "+rt.Token)
	}

	resp, err := rt.Transport.RoundTrip(req)
	if err != nil {
		rt.Logger.Error("Request failed: %v", err)
		return nil, err
	}

	// Если токен истек, обновляем его
	if resp.StatusCode == http.StatusUnauthorized {
		if rt.RefreshFn != nil {
			rt.Logger.Warn("Token expired, refreshing...")
			newToken, refreshErr := rt.RefreshFn(req.Context())
			if refreshErr != nil {
				rt.Logger.Error("Token refresh failed: %v", refreshErr)
				return nil, ErrTokenRefreshFailed
			}

			req.Header.Set("Authorization", "Bearer "+newToken)
			rt.Logger.Info("Token refreshed: %s", newToken)
			return rt.Transport.RoundTrip(req)
		}
		return resp, ErrTokenExpired
	}

	// Обработка 405 Method Not Allowed
	if resp.StatusCode == http.StatusMethodNotAllowed {
		rt.Logger.Error("Method Not Allowed (405): %s", req.URL.String())
		return nil, ErrMethodNotAllowed
	}

	// Обработка 404 Not Found
	if resp.StatusCode == http.StatusNotFound {
		rt.Logger.Error("Endpoint Not Found (404): %s", req.URL.String())
		return nil, ErrEndpointNotFound
	}

	return resp, nil
}
