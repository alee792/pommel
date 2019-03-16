// Package pommel is an S3ish Vault client that interacts with Paths and Keys
// as if Values are blob files.
// A typical use case is reading a JSON encoded file.
// Pommel does NOT provide authentication or retrieve a valid token.
package pommel

import (
	"context"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// Config for a Client.
type Config struct {
	// Address of the Vault Server.
	Addr string
	// Authentication token checked by Vault.
	Token string
}

// Client resolves secrets from Vault.
type Client struct {
	Config *Config
	*api.Client
}

// NewClient returns a default Client using credentials
// passed explicitly by the user.
func NewClient(cfg *Config) (*Client, error) {
	client, err := api.NewClient(&api.Config{
		Address: cfg.Addr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Could not create client. Are you logged in?")
	}
	c := &Client{
		Config: cfg,
		Client: client,
	}
	c.SetToken(c.Config.Token)
	return c, nil
}

// AutoConfig checks for Config fields in  a user's environment.
// There's no guarantee that this creates a well formed config.
func AutoConfig() (*Config, error) {
	return createConfig(Args{})
}

// Get value from Vault.
// ctx is unused because vault/api does not support it, but there's
// a medium chance that vault/api will be dropped in favor of a standard HTTP
// client to avoid a massive dependency graph.
func (c *Client) Get(ctx context.Context, bucket, key string) ([]byte, error) {
	secret, err := c.Logical().Read(bucket)
	if err != nil {
		return nil, err
	}
	v, ok := secret.Data[key]
	if !ok {
		return nil, errors.New("key does not exist")
	}
	return []byte(v.(string)), nil
}
