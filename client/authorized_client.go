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

func (c *AuthorizedDaichiClient) Ping(ctx context.Context) (string, error) {
	c.baseClient.Logger("[INFO] Pinging API...")
	return c.baseClient.Ping(ctx)
}
