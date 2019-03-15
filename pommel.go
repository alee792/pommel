// Package pommel is an S3ish Vault client that interacts with Paths and Keys
// as if Values are blob files.
// A typical use case is reading a JSON encoded file.
// Pommel does NOT provide authentication or retrieve a valid token.
package pommel

import (
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// Args from the command line.
type Args struct {
	Addr      string `arg:"-a" help:"vault addr"`
	TokenPath string `arg:"-p" help:"path to token"`
	Token     string `arg:"-t", help:"vault token:`
	Bucket    string `arg:"-b,required" help:"path to value"`
	Key       string `arg:"-k,required" help:"key for value"`
}

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
func (c *Client) Get(bucket, key string) ([]byte, error) {
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
