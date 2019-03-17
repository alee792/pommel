package main

import (
	"github.com/alee792/pommel/internal/cobra"
)

func main() {
	cmd := cobra.RootCmd()
	cmd.Execute()
}
