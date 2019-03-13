package main

import (
	"fmt"

	"git.target.com/ae-authentication/pommel"
)

func main() {
	pom := pommel.NewPommel("")
	raw, err := pom.ParseAndRead()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(raw))
}
