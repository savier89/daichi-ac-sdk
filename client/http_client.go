package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/savier89/circuitbreaker"
)

const (
	// ✅ Установлено строго по инструкции из DefaultAPIURL.txt
	DefaultAPIURL    = "https://web.daichicloud.ru/api/v4 "
	DefaultTokenPath = "/token"
	DefaultClientID  = "sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"
	DefaultRetries   = 3
)

type DaichiClient struct {
	clientID   string
	username   string
	password   string
	httpClient *http.Client
	token      string
	tokenMutex sync.RWMutex
	Logger     func(string, ...interface{})
	breaker    *circuitbreaker.CircuitBreaker
}

// Option — функциональный тип для настройки клиента
type Option func(*DaichiClient)

// WithClientID — устанавливает ClientID
func WithClientID(id string) Option {
	return func(c *DaichiClient) {
		c.clientID = id
	}
}

// WithUsername — устанавливает username
func WithUsername(username string) Option {
	return func(c *DaichiClient) {
		c.username = username
	}
}

// WithPassword — устанавливает password
func WithPassword(password string) Option {
	return func(c *DaichiClient) {
		c.password = password
	}
}

// WithRetries — устанавливает количество повторных попыток
func WithRetries(retries int) Option {
	return func(c *DaichiClient) {
		c.httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				DisableKeepAlives:     false,
				MaxIdleConnsPerHost:   50,
				ForceAttemptHTTP2:     true,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	}
}

// WithLogger — устанавливает пользовательский логгер
func WithLogger(logger func(string, ...interface{})) Option {
	return func(c *DaichiClient) {
		if logger == nil {
			logger = func(format string, args ...interface{}) {}
		}
		c.Logger = logger
	}
}

// WithDebug — включает или выключает DEBUG-логирование
func WithDebug(enable bool) Option {
	return func(c *DaichiClient) {
		if enable {
			c.Logger = func(format string, args ...interface{}) {
				log.Printf("[DEBUG] "+format, args...)
			}
		} else {
			c.Logger = func(format string, args ...interface{}) {}
		}
	}
}

// WithCircuitBreaker — устанавливает Circuit Breaker
func WithCircuitBreaker(b *circuitbreaker.CircuitBreaker) Option {
	return func(c *DaichiClient) {
		c.breaker = b
	}
}

// NewDaichiClient — создаёт клиент с опциями
func NewDaichiClient(opts ...Option) *DaichiClient {
	client := &DaichiClient{
		clientID: DefaultClientID,
		username: "",
		password: "",
		Logger:   func(format string, args ...interface{}) {},
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				DisableKeepAlives:     false,
				MaxIdleConnsPerHost:   50,
				ForceAttemptHTTP2:     true,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		token: "",
		breaker: NewCircuitBreaker(CircuitBreakerConfig{
			Name:        "daichi_api_breaker",
			MaxRequests: 5,
			Interval:    30 * time.Second,
			Timeout:     10 * time.Second,
			IsError: func(err error) bool {
				return err != nil
			},
		}),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// buildTokenRequest — создаёт POST-запрос для получения токена
func buildTokenRequest(ctx context.Context, grantType string, c *DaichiClient) (*http.Request, error) {
	values := url.Values{}
	values.Add("grant_type", grantType)
	values.Add("email", c.username)
	values.Add("password", c.password)
	values.Add("clientId", c.clientID)

	body := strings.NewReader(values.Encode())

	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), strings.TrimSpace(DefaultTokenPath))
	if err != nil {
		c.Logger("[ERROR] Failed to join token URL: %v", err)
		return nil, fmt.Errorf("invalid token URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, body)
	if err != nil {
		c.Logger("[ERROR] Failed to create token request: %v", err)
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	c.Logger("[INFO] Token request URL: %s", reqURL)
	c.Logger("[INFO] Token request body: %s", values.Encode())
	return req, nil
}

// fetchToken — общая логика получения токена
func (c *DaichiClient) fetchToken(ctx context.Context, req *http.Request) (string, error) {
	var resp *http.Response
	var err error

	for i := 0; i < DefaultRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
		c.Logger("[WARN] Retry #%d for %s", i+1, req.URL.String())
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		c.Logger("[ERROR] Token request failed: %v", err)
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Logger("[DEBUG] Token response: %s", string(body))

	if resp.StatusCode == http.StatusMethodNotAllowed {
		c.Logger("[ERROR] Method Not Allowed (405)")
		return "", ErrMethodNotAllowed
	}

	if resp.StatusCode == http.StatusNotFound {
		c.Logger("[ERROR] API endpoint not found (404): %s", req.URL.String())
		return "", ErrEndpointNotFound
	}

	if resp.StatusCode != http.StatusOK {
		c.Logger("[ERROR] Token request failed with status %d", resp.StatusCode)
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Done           bool `json:"done"`
		Errors         any  `json:"errors"`
		UpdateRequired bool `json:"updateRequired"`
		Data           struct {
			Token string `json:"access_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		c.Logger("[ERROR] Failed to decode token response: %v", err)
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if !result.Done {
		c.Logger("[ERROR] Token request failed: %v", result.Errors)
		return "", fmt.Errorf("token request failed: %v", result.Errors)
	}

	if result.UpdateRequired {
		return "", ErrTokenRefreshFailed
	}

	if result.Errors != nil {
		return "", fmt.Errorf("server returned errors: %v", result.Errors)
	}

	token := result.Data.Token
	if token == "" {
		return "", ErrTokenNotFound
	}

	return token, nil
}

// GetToken — получает токен авторизации
func (c *DaichiClient) GetToken(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		c.Logger("[ERROR] Username and password must be set")
		return ErrMissingCredentials
	}

	req, err := buildTokenRequest(ctx, "password", c)
	if err != nil {
		c.Logger("[ERROR] Failed to build token request: %v", err)
		return err
	}

	token, err := c.fetchToken(ctx, req)
	if err != nil {
		c.Logger("[ERROR] Failed to fetch token: %v", err)
		return err
	}

	c.tokenMutex.Lock()
	c.token = token
	c.tokenMutex.Unlock()

	c.Logger("[INFO] Token received: %s", token)
	return nil
}

// RefreshToken — обновляет токен
func (c *DaichiClient) RefreshToken(ctx context.Context) (string, error) {
	if c.username == "" || c.password == "" {
		c.Logger("[ERROR] Username and password must be set for token refresh")
		return "", ErrMissingCredentials
	}

	req, err := buildTokenRequest(ctx, "password", c)
	if err != nil {
		c.Logger("[ERROR] Failed to build refresh token request: %v", err)
		return "", err
	}

	token, err := c.fetchToken(ctx, req)
	if err != nil {
		c.Logger("[ERROR] Token refresh failed: %v", err)
		return "", fmt.Errorf("token refresh failed: %w", err)
	}

	c.tokenMutex.Lock()
	c.token = token
	c.tokenMutex.Unlock()

	c.Logger("[INFO] Token refreshed: %s", token)
	return token, nil
}

// Ping — проверяет доступность API
func (c *DaichiClient) Ping(ctx context.Context) (string, error) {
	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), "")
	if err != nil {
		c.Logger("[ERROR] Failed to build API URL: %v", err)
		return "", fmt.Errorf("invalid API URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		c.Logger("[ERROR] Failed to create ping request: %v", err)
		return "", fmt.Errorf("failed to create ping request: %w", err)
	}

	c.tokenMutex.RLock()
	token := c.token
	c.tokenMutex.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger("[ERROR] API unreachable: %v", err)
		return "", fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.Logger("[ERROR] API endpoint not found (404): %s", reqURL)
		return "", ErrEndpointNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger("[ERROR] Non-200 status code: %d", resp.StatusCode)
		return "", fmt.Errorf("non-200 status code: %d, response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger("[ERROR] Failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	c.Logger("[INFO] Ping successful: %s", string(body))
	return string(body), nil
}
