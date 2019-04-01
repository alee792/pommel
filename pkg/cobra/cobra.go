package cobra

import (
	"context"
	"fmt"
	"io"
	"os"

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
	root.PersistentFlags().BoolVarP(&hilt.HidePrompt, "hide", "h", false, "Hide prompt to print to stdout.")

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
			raw, err := h.Provider("vault").Client.Get(context.Background(), h.Bucket, h.Key)
			if err != nil {
				return errors.Wrap(err, "Get failed")
			}

			// Don't be clever...
			var in string
			if h.HidePrompt {
				io.Copy(os.Stdout, raw)
				return nil
			}

			// Show Prompt
			fmt.Println("Do you want to display this secret? (y/n)")
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
			return cli.ValidateSrcDst(h, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			srcS, srcB, srcK, err := cli.ParseURI(args[0])
			if err != nil {
				return errors.Wrap(err, "could not parse src")
			}
			dstS, dstB, dstK, err := cli.ParseURI(args[1])
			if err != nil {
				return errors.Wrap(err, "could not parse dst")
			}

			r, err := h.Provider(srcS).Client.Get(ctx, srcB, srcK)
			if err != nil {
				return errors.Wrap(err, "could not get from src")
			}
			if err := h.Provider(dstS).Client.Put(ctx, r, dstB, dstK); err != nil {
				return errors.Wrap(err, "could not put to dst")
			}
			return nil
		},
	}
	return cmd
}

func preRootAction(h *cli.Hilt) cmder {
	return func(cmd *cobra.Command, args []string) error {
		// Additional defaults.
		if h.Addr == "" {
			h.Addr = os.Getenv("VAULT_ADDR")
		}

		cfg, err := cli.CreateConfig(h.Flags)
		if err != nil {
			return errors.Wrap(err, "Config creation failed")
		}
		client, err := pommel.NewClient(cfg)
		if err != nil {
			return errors.Wrap(err, "Vault client creation failed")
		}

		h.AddProvider(&cli.Provider{
			Scheme: "vault",
			Client: client,
		})
		return nil
	}
}

func rootAction(h *cli.Hilt) cmder {
	return func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%+v\n%+v\n", h.Provider("vault").Client, h.Flags)
		fmt.Println(cmd.UsageString())
		return nil
	}
}
