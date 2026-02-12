//go:build ignore

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fezcode/gobake"
)

func main() {
	bake := gobake.NewEngine()
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		fmt.Printf("Error loading recipe.piml: %v\n", err)
		os.Exit(1)
	}

	bake.Task("test", "Run unit tests", func(ctx *gobake.Context) error {
		return ctx.Run("go", "test", "./...")
	})

	bake.Task("build", "Build the gobake CLI", func(ctx *gobake.Context) error {
		output := "gobake"
		if runtime.GOOS == "windows" {
			output = "gobake.exe"
		}
		ldflags := fmt.Sprintf("-X github.com/fezcode/gobake.Version=%s", bake.Info.Version)
		return ctx.Run("go", "build", "-ldflags", ldflags, "-o", output, "./cmd/gobake")
	})

	bake.TaskWithDeps("install", "Install gobake to GOPATH/bin", []string{"build"}, func(ctx *gobake.Context) error {
		return ctx.Run("go", "install", "./cmd/gobake")
	})

	bake.Execute()
}
