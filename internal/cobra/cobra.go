package cobra

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/alee792/pommel/pkg/cli"

	"github.com/alee792/pommel"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type cmder func(*cobra.Command, []string) error

// RootCmd returns a root command with sensible defaults.
func RootCmd() *cobra.Command {
	hilt := cli.NewHilt()

	root := &cobra.Command{
		Use:   "pommel",
		Short: "Pommel interacts with Vault as if it were a blob store",
		// Used to instantiate a client.
		PersistentPreRunE: preRootAction(hilt),
		RunE:              rootAction(hilt),
	}

	// Required flags. Will be replaced by args.
	root.PersistentFlags().StringVarP(&hilt.Bucket, "bucket", "b", "", "A path in Vault.")
	root.PersistentFlags().StringVarP(&hilt.Key, "key", "k", "", "A key in Vault.")
	root.MarkPersistentFlagRequired("bucket")
	root.MarkPersistentFlagRequired("key")

	// Optional flags.
	root.PersistentFlags().StringVarP(&hilt.Addr, "addr", "a", "", "Address of Vault server.")
	root.PersistentFlags().StringVarP(&hilt.Token, "tkn", "t", "", "Vault auth token.")
	root.PersistentFlags().StringVarP(&hilt.TokenPath, "tknp", "p", "~/.vault-token", "Path to Vault auth token.")

	// Subcommands.
	root.AddCommand(GetCmd(hilt))

	return root
}

// GetCmd sets up the cmd for a Pommel Get.
func GetCmd(h *cli.Hilt) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g", "read", "r"},
		Short:   "get value from Vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := h.Providers["vault"].Get(context.Background(), h.Bucket, h.Key)
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

// CpCmd sets up the cmd for copying files.
func CpCmd(h *cli.Hilt) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cp",
		Aliases: []string{"copy"},
		Short:   "copy files b/n locations",
		Args: func(cmd *cobra.Command, args []string) error {
			return validateSrcDst(h, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

// createConfig from Args and attempt to set default variables
// from a user's enivronment.
func createConfig(f *cli.Flags) (*pommel.Config, error) {
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

func preRootAction(h *cli.Hilt) cmder {
	return func(cmd *cobra.Command, args []string) error {
		// Additional defaults.
		if h.Addr == "" {
			h.Addr = os.Getenv("VAULT_ADDR")
		}

		cfg, err := createConfig(h.Flags)
		if err != nil {
			return errors.Wrap(err, "Config creation failed")
		}
		h.Handlers["vault"], err = pommel.NewClient(cfg)
		if err != nil {
			return errors.Wrap(err, "Client creation failed")
		}
		return nil
	}
}

func rootAction(h *cli.Hilt) cmder {
	return func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%+v\n%+v\n", h.Handlers["vault"], h.Flags)
		fmt.Println(cmd.UsageString())
		return nil
	}
}

// Either the soruce or destination location must be a valid URI.
// We're not in the business of local file managment here!
func validateSrcDst(h *cli.Hilt, args []string) error {
	if len(args) != 2 {
		return errors.New("requires exactly two args")
	}
	// Verbose logic for verbose errors.
	if !hasValidPrefix(args[0], h.Schemes) && !hasValidPrefix(args[1], h.Schemes) {
		return errors.New("requires valid URI")
	}
	return nil
}

func hasValidPrefix(s string, pp []string) bool {
	for _, p := range pp {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func parseURI(uri string) (schemes, bucket, key string, err error) {
	sep := "://"
	ss := strings.Split(uri, sep)
	if len(ss) != 2 {
		return "", "", "", errors.New("invalid uri")
	}
	scheme, path := ss[0], ss[1]

	bucket, key = filepath.Split(path)
	if bucket == "" || key == "" {
		return "", "", "", errors.New("bucket and key required")
	}
	return scheme, bucket, key, err
}
