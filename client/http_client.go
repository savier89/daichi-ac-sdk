package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/savier89/circuitbreaker"
)

// APIResponse — обертка для ответа сервера
type APIResponse[T any] struct {
	Done   bool `json:"done"`
	Errors any  `json:"errors"`
	Data   T    `json:"data"`
}

// Константы
const (
	DefaultAPIURL        = "https://web.daichicloud.ru/api/v4 "
	DefaultUserInfoPath  = "/user"
	DefaultBuildingsPath = "/buildings"
	DefaultTokenPath     = "/token"
	DefaultClientID      = "sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"
)

// DaichiClient — клиент для работы с API
type DaichiClient struct {
	clientID   string
	username   string
	password   string
	httpClient *http.Client
	token      string
	tokenMutex sync.RWMutex
	Logger     *Logger
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

// WithLogger — устанавливает пользовательский логгер
func WithLogger(logger *Logger) Option {
	return func(c *DaichiClient) {
		if logger == nil {
			logger = NewLogger(LogInfo, os.Stderr)
		}
		c.Logger = logger
	}
}

// WithLogLevel — устанавливает уровень логирования
func WithLogLevel(level LogLevel) Option {
	return func(c *DaichiClient) {
		if c.Logger == nil {
			c.Logger = NewLogger(level, os.Stderr)
		} else {
			c.Logger.SetLevel(level)
		}
	}
}

// WithCircuitBreaker — устанавливает Circuit Breaker
func WithCircuitBreaker(b *circuitbreaker.CircuitBreaker) Option {
	return func(c *DaichiClient) {
		c.breaker = b
	}
}

// WithDebug — включает дебаг-логи
func WithDebug(debug bool) Option {
	return func(c *DaichiClient) {
		if debug {
			c.Logger.SetLevel(LogDebug)
		} else {
			c.Logger.SetLevel(LogInfo)
		}
	}
}

// WithNoLogs — отключает все логи
func WithNoLogs() Option {
	return WithLogLevel(LogNone)
}

// NewDaichiClient — создает клиент с опциями
func NewDaichiClient(opts ...Option) *DaichiClient {
	client := &DaichiClient{
		clientID: DefaultClientID,
		username: "",
		password: "",
		Logger:   NewLogger(LogInfo, os.Stderr),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
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

// buildTokenRequest — создает POST-запрос для получения токена
func buildTokenRequest(ctx context.Context, c *DaichiClient) (*http.Request, error) {
	values := url.Values{
		"grant_type": {"password"},
		"email":      {c.username},
		"password":   {c.password},
		"clientId":   {c.clientID},
	}

	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), strings.TrimSpace(DefaultTokenPath))
	if err != nil {
		c.Logger.Error("Failed to build token URL: %v", err)
		return nil, fmt.Errorf("invalid token URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(values.Encode()))
	if err != nil {
		c.Logger.Error("Failed to create token request: %v", err)
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.URL.RawQuery = values.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	c.Logger.Debug("Token request URL: %s?%s", reqURL, req.URL.RawQuery)
	return req, nil
}

// fetchToken — общая логика получения токена
func (c *DaichiClient) fetchToken(ctx context.Context, req *http.Request) (string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger.Error("Token request failed: %v", err)
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("Failed to read token response: %v", err)
		return "", fmt.Errorf("failed to read token response: %w", err)
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
		c.Logger.Error("Failed to decode token response: %v", err)
		return "", fmt.Errorf("token unmarshal failed: %w", err)
	}

	if !result.Done {
		c.Logger.Error("Token request failed: %v", result.Errors)
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

	c.Logger.Info("Token received: %s", token)
	return token, nil
}

// GetToken — авторизация через /token
func (c *DaichiClient) GetToken(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		c.Logger.Error("Username and password must be set")
		return ErrMissingCredentials
	}

	req, err := buildTokenRequest(ctx, c)
	if err != nil {
		return err
	}

	token, err := c.fetchToken(ctx, req)
	if err != nil {
		c.Logger.Error("Failed to fetch token: %v", err)
		return err
	}

	c.tokenMutex.Lock()
	c.token = token
	c.tokenMutex.Unlock()

	return nil
}

// buildUserInfoRequest — создает GET-запрос для получения информации о пользователе
func buildUserInfoRequest(ctx context.Context, c *DaichiClient) (*http.Request, error) {
	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), strings.TrimSpace(DefaultUserInfoPath))
	if err != nil {
		c.Logger.Error("Failed to build user info URL: %v", err)
		return nil, fmt.Errorf("invalid user info URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		c.Logger.Error("Failed to create user info request: %v", err)
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	c.tokenMutex.RLock()
	token := c.token
	c.tokenMutex.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	req.Header.Set("Accept", "application/json")
	c.Logger.Debug("User info request URL: %s", reqURL)
	return req, nil
}

// GetUserInfo — возвращает информацию о пользователе
func (c *DaichiClient) GetUserInfo(ctx context.Context) (*DaichiUser, error) {
	req, err := buildUserInfoRequest(ctx, c)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger.Error("API unreachable: %v", err)
		return nil, fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.Logger.Error("API endpoint not found (404): %s", req.URL.String())
		return nil, ErrEndpointNotFound
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		c.Logger.Error("Method Not Allowed (405): %s", req.URL.String())
		return nil, ErrMethodNotAllowed
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("Non-200 status code: %d, response: %s", resp.StatusCode, body)
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("Failed to read user info response: %v", err)
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	c.Logger.Debug("User info raw: %s", body)

	// ✅ Десериализуем через APIResponse[DaichiUser]
	var response APIResponse[DaichiUser]
	if err := json.Unmarshal(body, &response); err != nil {
		c.Logger.Error("Failed to decode user info: %v", err)
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if !response.Done {
		c.Logger.Error("Server returned errors: %v", response.Errors)
		return nil, fmt.Errorf("server errors: %v", response.Errors)
	}

	c.Logger.Info("User info received: %s", string(body))
	return &response.Data, nil // ✅ Возвращаем данные из поля data
}

// buildBuildingsRequest — создает GET-запрос для получения зданий
func buildBuildingsRequest(ctx context.Context, c *DaichiClient) (*http.Request, error) {
	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), "buildings")
	if err != nil {
		c.Logger.Error("Failed to build buildings URL: %v", err)
		return nil, fmt.Errorf("invalid buildings URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		c.Logger.Error("Failed to create buildings request: %v", err)
		return nil, fmt.Errorf("failed to create buildings request: %w", err)
	}

	c.tokenMutex.RLock()
	token := c.token
	c.tokenMutex.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")
	c.Logger.Debug("Buildings request URL: %s", reqURL)

	return req, nil
}

// DaichiBuilding — структура здания с вложенными устройствами (экспортированная)
type DaichiBuilding struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Access      string `json:"access"`
	PlacesCount int    `json:"placesCount"`
	ShareCount  int    `json:"shareCount"`
	UTC         int    `json:"utc"`
	Coordinates struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinates"`
	GeoMode     bool                         `json:"geoMode"`
	GeoState    string                       `json:"geoState"`
	GeoZone     int                          `json:"geoZone"`
	Address     string                       `json:"address"`
	TriggeredBy *interface{}                 `json:"triggeredBy,omitempty"`
	HasSettings bool                         `json:"hasSettings"`
	OwnTrigger  *interface{}                 `json:"ownTrigger,omitempty"`
	CloudType   string                       `json:"cloudType"`
	TimeZone    string                       `json:"timeZone"`
	Image       string                       `json:"image"`
	Slogan      string                       `json:"slogan"`
	Places      []DaichiBuildingDeviceStruct `json:"places"` // ✅ Теперь структура
}

// GetBuildings — возвращает список зданий
func (c *DaichiClient) GetBuildings(ctx context.Context) ([]DaichiBuilding, error) {
	req, err := buildBuildingsRequest(ctx, c)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger.Error("API unreachable: %v", err)
		return nil, fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.Logger.Error("API endpoint not found (404): %s", req.URL.String())
		return nil, ErrEndpointNotFound
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		c.Logger.Error("Method Not Allowed (405): %s", req.URL.String())
		return nil, ErrMethodNotAllowed
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("Non-200 status code: %d, response: %s", resp.StatusCode, body)
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("Failed to read buildings response: %v", err)
		return nil, fmt.Errorf("failed to read buildings response: %w", err)
	}

	var response APIResponse[[]DaichiBuilding]
	if err := json.Unmarshal(body, &response); err != nil {
		c.Logger.Error("Failed to decode buildings: %v", err)
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if !response.Done {
		c.Logger.Error("Server returned errors: %v", response.Errors)
		return nil, fmt.Errorf("server errors: %v", response.Errors)
	}

	// Преобразуем []DaichiBuildingDeviceStruct → []Device
	for i := range response.Data {
		for j := range response.Data[i].Places {
			response.Data[i].Places[j] = DaichiBuildingDeviceStruct(response.Data[i].Places[j])
		}
	}

	c.Logger.Info("Buildings received: %d", len(response.Data))
	return response.Data, nil
}

// formatJSON — форматирует JSON для логирования
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		return fmt.Sprintf("failed to format JSON: %v", err)
	}
	return prettyJSON.String()
}

// formatDeviceState — форматирует состояние устройства для логирования
func formatDeviceState(device DaichiBuildingDeviceStruct) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ID: %d\n", device.ID))
	sb.WriteString(fmt.Sprintf("Title: %s\n", device.Title))
	sb.WriteString(fmt.Sprintf("Status: %s\n", device.Status))
	sb.WriteString(fmt.Sprintf("CurTemp: %.1f°C\n", device.CurTemp))
	sb.WriteString(fmt.Sprintf("IsOnline: %v\n", device.IsOnline()))
	sb.WriteString("State:\n")
	sb.WriteString(fmt.Sprintf("  IsOn: %v\n", device.State.IsOn))
	sb.WriteString("  Info:\n")
	sb.WriteString(fmt.Sprintf("    Text: %s\n", device.State.Info.Text))
	sb.WriteString("    Icons:\n")
	for i, icon := range device.State.Info.Icons {
		sb.WriteString(fmt.Sprintf("      %d: %s\n", i+1, icon))
	}
	sb.WriteString("    IconNames:\n")
	for i, name := range device.State.Info.IconNames {
		sb.WriteString(fmt.Sprintf("      %d: %s\n", i+1, name))
	}
	sb.WriteString("  Details:\n")
	for _, detail := range device.State.Details {
		for _, d := range detail.Details {
			if d.Text != nil {
				sb.WriteString(fmt.Sprintf("    Text: %s\n", *d.Text))
			}
			if d.Icon != nil {
				sb.WriteString(fmt.Sprintf("    Icon: %s\n", *d.Icon))
			}
		}
	}

	return sb.String()
}

// GetDeviceState — получает состояние устройства
func (c *DaichiClient) GetDeviceState(ctx context.Context, deviceID int) (*DaichiBuildingDeviceStruct, error) {
	// ✅ Исправленный URL: /devices/{id}, а не /devices/{id}
	devicePath := fmt.Sprintf("devices/%d", deviceID)
	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), strings.TrimSpace(devicePath))
	if err != nil {
		c.Logger.Error("Failed to build device URL: %v", err)
		return nil, fmt.Errorf("invalid device URL: %w", err)
	}

	// Создаем GET-запрос
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		c.Logger.Error("Failed to create device request: %v", err)
		return nil, fmt.Errorf("failed to create device request: %w", err)
	}

	c.tokenMutex.RLock()
	token := c.token
	c.tokenMutex.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")
	c.Logger.Debug("Device request URL: %s", reqURL)

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger.Error("API unreachable: %v", err)
		return nil, fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("Failed to read device response: %v", err)
		return nil, fmt.Errorf("failed to read device response: %w", err)
	}

	c.Logger.Debug("Device response raw: \n%s", formatJSON(body))
	c.Logger.Debug("Device response raw (escaped): %s", body)

	// Проверяем, что это JSON
	if !json.Valid(body) {
		c.Logger.Error("Invalid JSON response: %s", body)
		return nil, fmt.Errorf("invalid JSON response: %s", body)
	}

	// Десериализуем через APIResponse
	var response APIResponse[DaichiBuildingDeviceStruct]
	if err := json.Unmarshal(body, &response); err != nil {
		c.Logger.Error("Failed to decode device: %v", err)
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if !response.Done {
		c.Logger.Error("Server returned errors: %v", response.Errors)
		return nil, fmt.Errorf("server errors: %v", response.Errors)
	}

	// ✅ Улучшенный вывод состояния устройства
	c.Logger.Info("Device state received: \n%s", formatDeviceState(response.Data))
	return &response.Data, nil
}
