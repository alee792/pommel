package main

import (
	"fmt"
	"log"

	"github.com/alee792/pommel"
)

func main() {
	pom, args, err := pommel.CLI()
	if err != nil {
		log.Fatal(err)
	}
	raw, err := pom.Get(args.Bucket, args.Key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Do you want to display this secret? (y/n)")
	var in string
	fmt.Scanln(&in)
	if in == "y" {
		fmt.Println(string(raw))
	}
}
