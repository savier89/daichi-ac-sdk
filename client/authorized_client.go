package client

import (
	"context"
	"time"
)

// AuthorizedDaichiClient — клиент с авторизацией
type AuthorizedDaichiClient struct {
	*DaichiClient // ✅ Используем экспортированный тип
}

// NewAuthorizedDaichiClient — создает авторизованный клиент
func NewAuthorizedDaichiClient(ctx context.Context, username, password string, opts ...Option) (*AuthorizedDaichiClient, error) {
	opts = append(opts, WithUsername(username), WithPassword(password))
	client := NewDaichiClient(opts...)

	authCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client.Logger.Info("Authenticating...")
	if err := client.GetToken(authCtx); err != nil {
		client.Logger.Error("Authentication failed: %v", err)
		return nil, err
	}

	client.Logger.Info("Authorized client created")
	return &AuthorizedDaichiClient{
		DaichiClient: client,
	}, nil
}

// GetMqttUserInfo — возвращает информацию о пользователе
func (c *AuthorizedDaichiClient) GetMqttUserInfo(ctx context.Context) (*DaichiUser, error) {
	c.Logger.Info("Fetching MQTT user info...")
	return c.DaichiClient.GetUserInfo(ctx)
}

// GetBuildings — возвращает список зданий
func (c *AuthorizedDaichiClient) GetBuildings(ctx context.Context) ([]DaichiBuilding, error) {
	c.Logger.Info("Fetching buildings...")
	return c.DaichiClient.GetBuildings(ctx)
}

// GetDeviceState — возвращает состояние устройства
func (c *AuthorizedDaichiClient) GetDeviceState(ctx context.Context, deviceID int) (*DaichiBuildingDeviceStruct, error) {
	c.Logger.Info("Fetching device state: %d", deviceID)
	return c.DaichiClient.GetDeviceState(ctx, deviceID)
}
