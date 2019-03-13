package main

import (
	"fmt"

	"github.com/alee792/pommel"
)

func main() {
	pom, args, err := pommel.CLI()
	if err != nil {
		panic(err)
	}
	raw, err := pom.Read(args.Path, args.Key)
	if err != nil {
		panic(err)
	}
	fmt.Println("Do you want to display this secret? (y/n)")
	var in string
	fmt.Scanln(&in)
	if in == "y" {
		fmt.Println(string(raw))
	}
}
