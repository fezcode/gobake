

package main

import (
	"fmt"
	"runtime"

	"github.com/fezcode/gobake"
)

func Run(bake *gobake.Engine) error {
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		return fmt.Errorf("Error loading recipe.piml: %v", err)
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

	bake.Task("tag", "Tag the current version", func(ctx *gobake.Context) error {
		tagName := "v" + bake.Info.Version
		ctx.Log("Creating git tag: %s", tagName)
		return ctx.Run("git", "tag", tagName)
	})

	bake.TaskWithDeps("release", "Tag and push to remote", []string{"tag"}, func(ctx *gobake.Context) error {
		tagName := "v" + bake.Info.Version
		ctx.Log("Pushing changes and tag %s...", tagName)
		if err := ctx.Run("git", "push"); err != nil {
			return err
		}
		return ctx.Run("git", "push", "origin", tagName)
	})

	return nil
}
