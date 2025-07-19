package client

import (
	"context"
	"net/http"
)

// AuthRoundTripper добавляет токен авторизации к каждому запросу
type AuthRoundTripper struct {
	Transport http.RoundTripper
	Token     string
	RefreshFn func(context.Context) (string, error)
	Logger    func(string, ...interface{})
}

// RoundTrip реализует интерфейс http.RoundTripper
func (rt *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.Token != "" {
		clonedReq := req.Clone(req.Context())
		clonedReq.Header.Set("Authorization", "Bearer "+rt.Token)
		req = clonedReq
	}

	resp, err := rt.Transport.RoundTrip(req)
	if err != nil {
		rt.Logger("[ERROR] Failed to execute request: %v", err)
		return nil, err
	}

	// Если токен истек, обновляем его
	if resp.StatusCode == http.StatusUnauthorized {
		if rt.RefreshFn != nil {
			rt.Logger("[INFO] Token expired, refreshing...")
			newToken, refreshErr := rt.RefreshFn(req.Context())
			if refreshErr != nil {
				rt.Logger("[ERROR] Failed to refresh token: %v", refreshErr)
				return nil, ErrTokenRefreshFailed
			}

			// Обновляем токен и повторяем запрос
			rt.Token = newToken
			rt.Logger("[INFO] Token refreshed: %s", newToken)

			// Создаем новый запрос с новым токеном
			newReq := req.Clone(req.Context())
			newReq.Header.Set("Authorization", "Bearer "+newToken)
			return rt.Transport.RoundTrip(newReq)
		}
		return resp, ErrTokenExpired
	}

	// Проверка на 405 Method Not Allowed
	if resp.StatusCode == http.StatusMethodNotAllowed {
		rt.Logger("[ERROR] Method Not Allowed (405): %s", req.URL.String())
		return nil, ErrMethodNotAllowed
	}

	// Проверка на 404 Not Found
	if resp.StatusCode == http.StatusNotFound {
		rt.Logger("[ERROR] Endpoint Not Found (404): %s", req.URL.String())
		return nil, ErrEndpointNotFound
	}

	return resp, nil
}
