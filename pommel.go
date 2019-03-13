// Package pommel extracts a JSON encoded byte slice from Vault.
// Pommel does NOT provide authentication or a valid token.
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
	Path      string `arg:"-p,required" help:"path to value"`
	Key       string `arg:"-k,required" help:"key for value"`
}

// Client resolves secrets from Vault.
type Client struct {
	Args   *Args
	Token  string
	Prefix string
	*api.Client
}

// NewClient instantiates a client.
func NewClient() *Client {
	a := new(Args)
	arg.MustParse(a)
	a.TokenPath = "~/.vault-token"
	return &Client{
		Args: a,
	}
}

// ParseAndRead is convenient if no custom CL args are needed.
func (c *Client) ParseAndRead() ([]byte, error) {
	if err := c.Parse(); err != nil {
		return nil, err
	}
	return c.Read()
}

// Parse command-line args and create Vault Client.
func (c *Client) Parse() error {
	if c.Args.Addr == "" {
		c.Args.Addr = os.Getenv("VAULT_ADDR")
	}

	// Expand "~" to absolute path.
	if strings.Contains(c.Args.TokenPath, "~") {
		usr, _ := user.Current()
		c.Args.TokenPath = strings.Replace(c.Args.TokenPath, "~", usr.HomeDir, -1)
	}
	tkn, err := ioutil.ReadFile(c.Args.TokenPath)
	if err != nil {
		return errors.Wrapf(err, "invalid token path %s", c.Args.TokenPath)
	}
	c.Token = string(tkn)
	client, err := api.NewClient(&api.Config{
		Address: c.Args.Addr,
	})
	if err != nil {
		return errors.Wrap(err, "Could not create client. Are you logged in?")
	}
	c.Client = client
	return nil
}

// Read from Vault.
func (c *Client) Read() ([]byte, error) {
	c.SetToken(c.Token)
	secret, err := c.Logical().Read(c.Args.Path)
	if err != nil {
		return nil, err
	}
	v, ok := secret.Data[c.Args.Key]
	if !ok {
		return nil, errors.New("key does not exist")
	}
	return []byte(v.(string)), nil
}

func (c *Client) validate() error {
	if c.Args.Path == "" {
		return errors.New("path is required")
	}
	if c.Args.Key == "" {
		return errors.New("key is required")
	}
	return nil
}
