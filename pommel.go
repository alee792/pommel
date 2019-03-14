// Package pommel is an S3ish Vault client that interacts with Paths and Keys
// as if Values are blob files.
// A typical use case is reading a JSON encoded file.
// Pommel does NOT provide authentication or retrieve a valid token.
package pommel

import (
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// Args from the command line.
type Args struct {
	Addr      string `arg:"-a" help:"vault addr"`
	TokenPath string `arg:"-t" help:"path to token"`
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

// CLI creates a Client using command line args and
// credentials found in a user's environment.
func CLI() (*Client, *Args, error) {
	a := new(Args)
	arg.MustParse(a)

	// Defaults
	a.TokenPath = "~/.vault-token"

	cfg, err := createConfig(*a)
	if err != nil {
		return nil, nil, err
	}
	c, err := NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return c, a, nil
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

func parseAddr(addr string) (string, error) {
	if addr == "" {
		addr = os.Getenv("VAULT_ADDR")
	}
	return addr, nil
}

func parseToken(tokenPath string) (string, error) {
	// Expand "~" to absolute path.
	if strings.Contains(tokenPath, "~") {
		usr, _ := user.Current()
		tokenPath = strings.Replace(tokenPath, "~", usr.HomeDir, -1)
	}
	tkn, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return "", errors.Wrapf(err, "invalid token path %s", tokenPath)
	}
	return string(tkn), nil
}

func createConfig(a Args) (*Config, error) {
	tkn, err := parseToken(a.TokenPath)
	if err != nil {
		return nil, err
	}
	addr, err := parseAddr(a.Addr)
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		Addr:  addr,
		Token: tkn,
	}
	return cfg, nil
}
