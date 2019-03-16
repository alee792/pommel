package main

import (
	"fmt"

	"github.com/alee792/pommel/internal/cli/cobra"
)

func main() {
	cmd, a := cobra.NewRootCommand()
	fmt.Printf("%+v\n", a)
	cmd.Execute()
}
