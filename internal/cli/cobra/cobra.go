package cobra

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/pkg/errors"

	"github.com/alee792/pommel"

	"github.com/spf13/cobra"
)

// Pommeler defines a Vault clients expected
// capabilities in an S3-like interface.
type Pommeler interface {
	Get(ctx context.Context, bucket, key string) (io.Reader, error)
	// Put(ctx context.Context, bucket, key string, body io.Reader) error
}

// Hilt allows Pommel and Cobra dependencies to be shared and reused.
type Hilt struct {
	Pommeler
	*Flags
}

type cmder func(*cobra.Command, []string) error

// Flags from the CLI.
type Flags struct {
	Addr      string `arg:"-a" help:"vault addr"`
	TokenPath string `arg:"-p" help:"path to token"`
	Token     string `arg:"-t" help:"vault token"`
	Bucket    string `arg:"-b,required" help:"path to value"`
	Key       string `arg:"-k,required" help:"key for value"`
}

// RootCmd returns a root command with sensible defaults.
func RootCmd() *cobra.Command {
	hilt := Hilt{
		Pommeler: &pommel.Client{},
		Flags:    &Flags{},
	}

	root := &cobra.Command{
		Use:   "pommel",
		Short: "Pommel interacts with Vault as if it were a blob store",
		// Used to instantiate a client.
		PersistentPreRunE: hilt.preRootAction(),
		RunE:              hilt.rootAction(),
	}

	// Required flags.
	root.PersistentFlags().StringVarP(&hilt.Bucket, "bucket", "b", "", "A path in Vault.")
	root.PersistentFlags().StringVarP(&hilt.Key, "key", "k", "", "A key in Vault.")
	root.MarkPersistentFlagRequired("bucket")
	root.MarkPersistentFlagRequired("key")

	// Optional flags.
	root.PersistentFlags().StringVarP(&hilt.Addr, "addr", "a", "", "Address of Vault server.")
	root.PersistentFlags().StringVarP(&hilt.Token, "tkn", "t", "", "Vault auth token.")
	root.PersistentFlags().StringVarP(&hilt.TokenPath, "tknp", "p", "~/.vault-token", "Path to Vault auth token.")

	// Subcommands.
	root.AddCommand(hilt.GetCmd())

	return root
}

// GetCmd value from Vault.
func (h *Hilt) GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g", "read", "r"},
		Short:   "get value from Vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := h.Get(context.Background(), h.Bucket, h.Key)
			if err != nil {
				return errors.Wrap(err, "Get failed")
			}
			fmt.Println("Do you want to display this secret? (y/n)")
			var in string
			fmt.Scanln(&in)
			if in == "y" {
				io.Copy(os.Stdout, raw)
			}
			return nil
		},
	}
	return cmd
}

// createConfig from Args and attempt to set default variables
// from a user's enivronment.
func createConfig(f *Flags) (*pommel.Config, error) {
	if f.TokenPath == "" {
		f.TokenPath = "~/.vault-token"
	}

	if f.Token == "" {
		tkn, err := getToken(f.TokenPath)
		if err != nil {
			return nil, err
		}
		f.Token = tkn
	}

	if f.Addr == "" {
		f.Addr = os.Getenv("VAULT_ADDR")
	}
	cfg := &pommel.Config{
		Addr:  f.Addr,
		Token: f.Token,
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

func (h *Hilt) preRootAction() cmder {
	return func(cmd *cobra.Command, args []string) error {
		// Additional defaults.
		if h.Addr == "" {
			h.Addr = os.Getenv("VAULT_ADDR")
		}

		cfg, err := createConfig(h.Flags)
		if err != nil {
			return errors.Wrap(err, "Config creation failed")
		}
		h.Pommeler, err = pommel.NewClient(cfg)
		if err != nil {
			return errors.Wrap(err, "Client creation failed")
		}
		return nil
	}
}

func (h *Hilt) rootAction() cmder {
	return func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%+v\n%+v\n", h.Pommeler, h.Flags)
		fmt.Println(cmd.UsageString())
		return nil
	}
}
