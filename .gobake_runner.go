package main

import (
	"fmt"
	"os"
	"github.com/fezcode/gobake"
)

func main() {
	bake := gobake.NewEngine()
	if err := Run(bake); err != nil {
		fmt.Fprintf(os.Stderr, "Recipe setup failed: %v\n", err)
		os.Exit(1)
	}
	bake.Execute()
}
