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

// MQTTUser — отдельная структура для MQTT-данных
type MQTTUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DeviceState — структура для поля state
type DeviceState struct {
	IsOn bool `json:"isOn"`
	Info struct {
		Text      string   `json:"text"`
		Icons     []string `json:"icons"`
		IconsSvg  []string `json:"iconsSvg"`
		IconNames []string `json:"iconNames"`
	} `json:"info"`
	Details []interface{} `json:"details"`
}

// DaichiBuildingDevice — структура устройства в здании
type DaichiBuildingDevice struct {
	ID                int             `json:"id"`
	Serial            string          `json:"serial"`
	Status            string          `json:"status"`
	Title             string          `json:"title"`
	CurTemp           float64         `json:"curTemp"`
	State             DeviceState     `json:"state"` // ✅ Теперь структура
	Features          map[string]bool `json:"features"`
	GroupID           *interface{}    `json:"groupId,omitempty"`
	BuildingID        int             `json:"buildingId"`
	LastOnline        string          `json:"lastOnline"`
	CreatedAt         string          `json:"createdAt"`
	Pinned            bool            `json:"pinned"`
	Access            string          `json:"access"`
	Progress          *interface{}    `json:"progress,omitempty"`
	CurrentPreset     *interface{}    `json:"currentPreset,omitempty"`
	Timer             *interface{}    `json:"timer,omitempty"`
	CloudType         string          `json:"cloudType"`
	DistributionType  string          `json:"distributionType"`
	Company           string          `json:"company"`
	IsBle             bool            `json:"isBle"`
	DeviceControlType string          `json:"deviceControlType"`
	FirmwareType      string          `json:"firmwareType"`
	VrfTitle          *interface{}    `json:"vrfTitle,omitempty"`
	DeviceType        string          `json:"deviceType"`
	Subscription      *interface{}    `json:"subscription,omitempty"`
	SubscriptionID    *int            `json:"subscriptionId,omitempty"`
	WarrantyNumber    *string         `json:"warrantyNumber,omitempty"`
	ConditionerSerial *string         `json:"conditionerSerial,omitempty"`
	UpdatedAt         *string         `json:"updatedAt,omitempty"`
	Online            bool            `json:"online,omitempty"`
}

// DaichiUser — структура данных пользователя
type DaichiUser struct {
	ID                       int           `json:"id"`
	Token                    string        `json:"token"`
	Email                    string        `json:"email"`
	MQTTUser                 *MQTTUser     `json:"mqttUser"`
	IsEmailConfirmed         bool          `json:"isEmailConfirmed"`
	Phone                    *string       `json:"phone,omitempty"`
	IsPhoneConfirmed         bool          `json:"isPhoneConfirmed"`
	FIO                      string        `json:"fio"`
	Company                  string        `json:"company"`
	UserType                 string        `json:"userType"`
	ExpiredIn                *string       `json:"expiredIn,omitempty"`
	DeleteAccountRequestedAt *string       `json:"deleteAccountRequestedAt,omitempty"`
	Image                    *string       `json:"image,omitempty"`
	AccessRequests           []interface{} `json:"accessRequests"`
}

// DaichiBuilding — структура здания с вложенными устройствами
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
	GeoMode     bool                   `json:"geoMode"`
	GeoState    string                 `json:"geoState"`
	GeoZone     int                    `json:"geoZone"`
	Address     string                 `json:"address"`
	TriggeredBy *interface{}           `json:"triggeredBy,omitempty"`
	HasSettings bool                   `json:"hasSettings"`
	OwnTrigger  *interface{}           `json:"ownTrigger,omitempty"`
	CloudType   string                 `json:"cloudType"`
	TimeZone    string                 `json:"timeZone"`
	Image       string                 `json:"image"`
	Slogan      string                 `json:"slogan"`
	Places      []DaichiBuildingDevice `json:"places"` // ✅ Список устройств
}

const (
	// ✅ Константа установлена строго по инструкции из DefaultAPIURL.txt
	DefaultAPIURL        = "https://web.daichicloud.ru/api/v4 "
	DefaultUserInfoPath  = "/user"
	DefaultBuildingsPath = "/buildings"
	DefaultTokenPath     = "/token"
	DefaultClientID      = "sOJO7B6SqgaKudTfCzqLAy540cCuDzpI"
	DefaultRetries       = 3
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

// WithDebug — включает или выключает DEBUG-логи
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
		Logger:   func(format string, args ...interface{}) {}, // по умолчанию — выключено
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

// buildUserInfoRequest — создаёт GET-запрос для получения информации о пользователе
func buildUserInfoRequest(ctx context.Context, c *DaichiClient) (*http.Request, error) {
	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), strings.TrimSpace(DefaultUserInfoPath))
	if err != nil {
		c.Logger("[ERROR] Failed to join user info URL: %v", err)
		return nil, fmt.Errorf("invalid user info URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		c.Logger("[ERROR] Failed to create user info request: %v", err)
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	c.tokenMutex.RLock()
	token := c.token
	c.tokenMutex.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	req.Header.Set("Accept", "application/json")
	c.Logger("[INFO] User info request URL: %s", reqURL)
	return req, nil
}

// GetUserInfo — возвращает информацию о пользователе в формате DaichiUser
func (c *DaichiClient) GetUserInfo(ctx context.Context) (*DaichiUser, error) {
	req, err := buildUserInfoRequest(ctx, c)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger("[ERROR] API unreachable: %v", err)
		return nil, fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.Logger("[ERROR] API endpoint not found (404): %s", req.URL.String())
		return nil, ErrEndpointNotFound
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		c.Logger("[ERROR] Method Not Allowed (405): %s", req.URL.String())
		return nil, ErrMethodNotAllowed
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger("[ERROR] Non-200 status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("non-200 status code: %d, response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger("[ERROR] Failed to read user info response: %v", err)
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	// Обертка вокруг data
	var result struct {
		Done   bool       `json:"done"`
		Errors any        `json:"errors"`
		Data   DaichiUser `json:"data"`
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		c.Logger("[ERROR] Failed to decode user info response: %v", err)
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	if !result.Done {
		c.Logger("[ERROR] Server returned errors: %v", result.Errors)
		return nil, fmt.Errorf("server returned errors: %v", result.Errors)
	}

	c.Logger("[INFO] User info received: %s", string(body))
	return &result.Data, nil
}

// buildBuildingsRequest — создаёт GET-запрос для получения зданий
func buildBuildingsRequest(ctx context.Context, c *DaichiClient) (*http.Request, error) {
	reqURL, err := url.JoinPath(strings.TrimSpace(DefaultAPIURL), strings.TrimSpace(DefaultBuildingsPath))
	if err != nil {
		c.Logger("[ERROR] Failed to join buildings URL: %v", err)
		return nil, fmt.Errorf("invalid buildings URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		c.Logger("[ERROR] Failed to create buildings request: %v", err)
		return nil, fmt.Errorf("failed to create buildings request: %w", err)
	}

	c.tokenMutex.RLock()
	token := c.token
	c.tokenMutex.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	req.Header.Set("Accept", "application/json")
	c.Logger("[INFO] Buildings request URL: %s", reqURL)
	return req, nil
}

// GetBuildings — возвращает список зданий
func (c *DaichiClient) GetBuildings(ctx context.Context) ([]DaichiBuilding, error) {
	req, err := buildBuildingsRequest(ctx, c)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger("[ERROR] API unreachable: %v", err)
		return nil, fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.Logger("[ERROR] API endpoint not found (404): %s", req.URL.String())
		return nil, ErrEndpointNotFound
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		c.Logger("[ERROR] Method Not Allowed (405): %s", req.URL.String())
		return nil, ErrMethodNotAllowed
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger("[ERROR] Non-200 status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("non-200 status code: %d, response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger("[ERROR] Failed to read buildings response: %v", err)
		return nil, fmt.Errorf("failed to read buildings response: %w", err)
	}

	// Обертка вокруг data
	var result struct {
		Done   bool             `json:"done"`
		Errors any              `json:"errors"`
		Data   []DaichiBuilding `json:"data"`
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		c.Logger("[ERROR] Failed to decode buildings response: %v", err)
		return nil, fmt.Errorf("failed to decode buildings response: %w", err)
	}

	if !result.Done {
		c.Logger("[ERROR] Server returned errors: %v", result.Errors)
		return nil, fmt.Errorf("server returned errors: %v", result.Errors)
	}

	c.Logger("[INFO] Buildings received: %s", string(body))
	return result.Data, nil
}
