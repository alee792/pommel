// Package pommel extracts a JSON encoded byte slice from Vault.
package pommel

import (
	"flag"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/peterbourgon/ff"
	"github.com/pkg/errors"
)

// Pommel resolves secrets from Vault.
type Pommel struct {
	FlagSet   *flag.FlagSet
	Addr      string `arg:"-a" help:"vault addr"`
	Token     string
	TokenPath string `arg:"-t" help:"path to token"`
	Path      string `arg:"-p" help:"path to value"`
	Key       string `arg:"-k,required" help:"key for value"`
	Prefix    string
	*api.Client
}

// NewPommel allows clients to insert their own flags before parsing.
func NewPommel(prefix string) *Pommel {
	p := new(Pommel)
	p.Prefix = prefix
	arg.MustParse(&p)
	return p
}

// ParseAndRead is convenient if no custom CL args are needed.
func (p *Pommel) ParseAndRead() ([]byte, error) {
	if err := p.Parse(); err != nil {
		return nil, err
	}
	return p.Read()
}

// Parse command-line args and create Vault Client.
func (p *Pommel) Parse() error {
	opts := []ff.Option{
		ff.WithConfigFileFlag("path"),
		ff.WithConfigFileParser(ff.JSONParser),
	}
	if p.Prefix != "" {
		opts = append(opts, ff.WithEnvVarPrefix(p.Prefix))
	}
	err := ff.Parse(p.FlagSet, os.Args[1:], opts...)
	if err != nil {
		return err
	}

	if err := p.validate(); err != nil {
		return err
	}

	// If no address is given, localhost will be used.
	if p.Addr == "" {
		p.Addr = os.Getenv("VAULT_ADDR")
	}

	// Expand "~" to absolute path.
	if strings.Contains(p.TokenPath, "~") {
		usr, _ := user.Current()
		p.TokenPath = strings.Replace(p.TokenPath, "~", usr.HomeDir, -1)
	}
	tkn, err := ioutil.ReadFile(p.TokenPath)
	if err != nil {
		return errors.Wrap(err, "invalid token path")
	}
	p.Token = string(tkn)
	client, err := api.NewClient(&api.Config{
		Address: p.Addr,
	})
	if err != nil {
		return errors.Wrap(err, "Could not create client. Are you logged in?")
	}
	p.Client = client
	return nil
}

// Read from Vault.
func (p *Pommel) Read() ([]byte, error) {
	p.SetToken(p.Token)
	secret, err := p.Logical().Read(p.Path)
	if err != nil {
		return nil, err
	}
	v := secret.Data[p.Key]
	return v.([]byte), nil
}

func (p *Pommel) validate() error {
	if p.Path == "" {
		return errors.New("path is required")
	}
	if p.Key == "" {
		return errors.New("key is required")
	}
	return nil
}
