package main

import (
	"github.com/alee792/pommel/internal/cli/cobra"
)

func main() {
	cmd := cobra.RootCmd()
	cmd.Execute()
}
