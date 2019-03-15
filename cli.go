package pommel

import (
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/pkg/errors"
)

// CLI creates a Client using command line args and
// credentials found in a user's environment.
func CLI() (*Client, *Args, error) {
	a := new(Args)
	arg.MustParse(a)

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

// createConfig from Args and attempt to set default variables
// from a user's enivronment.
func createConfig(a Args) (*Config, error) {
	if a.TokenPath == "" {
		a.TokenPath = "~/.vault-token"
	}

	if a.Token == "" {
		tkn, err := getToken(a.TokenPath)
		if err != nil {
			return nil, err
		}
		a.Token = tkn
	}

	if a.Addr == "" {
		a.Addr = os.Getenv("VAULT_ADDR")
	}
	cfg := &Config{
		Addr:  a.Addr,
		Token: a.Token,
	}
	return cfg, nil
}

func getToken(tokenPath string) (string, error) {
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
