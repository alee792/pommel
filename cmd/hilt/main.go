package main

import (
	"github.com/alee792/pommel/pkg/cobra"
)

func main() {
	cmd := cobra.RootCmd()
	cmd.Execute()
}
