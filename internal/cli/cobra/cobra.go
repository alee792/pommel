package cobra

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Pommeler defines a Vault clients expected
// capabilities in an S3-like interface.
type Pommeler interface {
	Get(ctx context.Context, bucket, key string) (io.Reader, error)
	Put(ctx context.Context, bucket, key string, body io.Reader) error
}

// Args from the CLI.
type Args struct {
	Addr      string `arg:"-a" help:"vault addr"`
	TokenPath string `arg:"-p" help:"path to token"`
	Token     string `arg:"-t" help:"vault token"`
	Bucket    string `arg:"-b,required" help:"path to value"`
	Key       string `arg:"-k,required" help:"key for value"`
}

// NewRootCommand returns a root command with sensible defaults.
func NewRootCommand() (*cobra.Command, *Args) {
	var a *Args
	root := &cobra.Command{
		Use:   "pommel",
		Short: "Pommel interacts with Vault as if it were a blob store",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Pommel")
		},
	}
	root.PersistentFlags().StringVarP(&a.Addr, "addr", "a", "", "Address of Vault server.")
	root.PersistentFlags().StringVarP(&a.Token, "tkn", "t", "", "Vault auth token.")
	root.PersistentFlags().StringVarP(&a.TokenPath, "tknp", "p", "~/.vault-token", "Path to Vault auth token.")
	root.MarkFlagFilename("tknp")

	root.PersistentFlags().StringVarP(&a.Bucket, "bucket", "b", "", "A path in Vault.")
	root.PersistentFlags().StringVarP(&a.Key, "key", "k", "", "A key in Vault.")
	root.MarkFlagRequired("bucket")
	root.MarkFlagRequired("key")

	// Additional defaults.
	if a.Addr == "" {
		a.Addr = os.Getenv("VAULT_ADDR")
	}

	return root, a
}
