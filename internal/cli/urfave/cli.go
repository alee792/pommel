// Package urfave is a urfave/cli implementation of Pommel.
// It's committed for now, but will be removed and replaced with Cobra.
package urfave

import (
	"context"
	"io"

	cli "gopkg.in/urfave/cli.v2"
)

// Pommeler defines a Vault clients expected
// capabilities in an S3-like interface.
type Pommeler interface {
	Get(ctx context.Context, bucket, key string) (io.Reader, error)
	Put(ctx context.Context, bucket, key string, body io.Reader) error
}

// App is a CLI for Pommel.
type App struct {
	app *cli.App
	p   *Pommeler
}

// Flags used by CLI.
var (
	AddrFlag = cli.StringFlag{
		Name:  "addr",
		Usage: "server addr",
	}
	TokenPathFlag = cli.StringFlag{
		Name:    "tokenPath",
		Usage:   "path to auth token",
		Value:   "~/.vault-token",
		Aliases: []string{"tknPath, tp, p"},
	}
	TokenFlag = cli.StringFlag{
		Name:    "token",
		Usage:   "auth token",
		Aliases: []string{"tkn, t"},
	}
	BucketFlag = cli.StringFlag{
		Name:    "bucket",
		Usage:   "path to key",
		Aliases: []string{"b"},
	}
	KeyPath = cli.StringFlag{
		Name:    "key",
		Usage:   "key of value",
		Aliases: []string{"k"},
	}
)

// NewApp sets up the CLI and provides sensible defaults.
func NewApp(p *Pommeler) *App {
	var app *cli.App
	app.Commands = []*cli.Command{
		{
			Name:    "get",
			Aliases: []string{"g", "read", "r"},
			Usage:   "get value from a bucket/key",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}
	a := &App{
		app: app,
	}
	return a
}
