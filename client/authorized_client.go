package client

import (
	"context"
	"time"
)

type AuthorizedDaichiClient struct {
	baseClient *DaichiClient
}

func NewAuthorizedDaichiClient(ctx context.Context, username, password string, opts ...Option) (*AuthorizedDaichiClient, error) {
	opts = append(opts, WithUsername(username), WithPassword(password))
	client := NewDaichiClient(opts...)

	authCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client.Logger("[INFO] Authenticating...")
	if err := client.GetToken(authCtx); err != nil {
		client.Logger("[ERROR] Authentication failed: %v", err)
		return nil, err
	}

	client.Logger("[INFO] Authorized client created")
	return &AuthorizedDaichiClient{
		baseClient: client,
	}, nil
}

// GetMqttUserInfo — возвращает информацию о пользователе в формате DaichiUser
func (c *AuthorizedDaichiClient) GetMqttUserInfo(ctx context.Context) (*DaichiUser, error) {
	c.baseClient.Logger("[INFO] Fetching MQTT user info...")
	return c.baseClient.GetUserInfo(ctx)
}

// GetBuildings — возвращает список зданий
func (c *AuthorizedDaichiClient) GetBuildings(ctx context.Context) ([]DaichiBuilding, error) {
	c.baseClient.Logger("[INFO] Fetching buildings...")
	return c.baseClient.GetBuildings(ctx)
}
