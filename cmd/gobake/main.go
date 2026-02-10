package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Handle "init" command
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runInit()
		return
	}

	// Check for Recipe.go
	if _, err := os.Stat("Recipe.go"); err == nil {
		// Found Recipe.go, run it
		// We pass all args to the Recipe.go
		args := append([]string{"run", "Recipe.go"}, os.Args[1:]...)
		cmd := exec.Command("go", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	// If no Recipe.go, check for recipe.piml to give a hint
	if _, err := os.Stat("recipe.piml"); err == nil {
		fmt.Println("Found recipe.piml but no Recipe.go.")
		fmt.Println("Please create a Recipe.go file to define your build tasks.")
		fmt.Println("\nExample Recipe.go:")
		fmt.Println(`package main

import "github.com/fezcode/gobake"

func main() {
	bake := gobake.NewEngine()
	bake.LoadRecipeInfo("recipe.piml")
	
	bake.Task("build", "Build the project", func(ctx *gobake.Context) error {
		return ctx.Run("go", "build", "-o", "bin/app")
	})
	
	bake.Execute()
}`)
		return
	}

	fmt.Println("gobake: No Recipe.go or recipe.piml found in the current directory.")
	fmt.Println("Run 'gobake init' to create a new project configuration.")
	fmt.Println("Visit https://github.com/fezcode/gobake for more information.")
}

func runInit() {
	if _, err := os.Stat("recipe.piml"); err == nil {
		fmt.Println("Error: recipe.piml already exists.")
		return
	}
	if _, err := os.Stat("Recipe.go"); err == nil {
		fmt.Println("Error: Recipe.go already exists.")
		return
	}

	// Default PIML content
	pimlContent := `(name) my-awesome-project
(version) 0.1.0
(description) A new project built with gobake
(license) MIT
(repository) github.com/user/my-awesome-project
(authors)
    - Me <me@example.com>
(keywords)
    - cli
    - tool
    - golang`

	// Default Recipe.go content
	recipeGoContent := `package main

import (
	"fmt"
	"github.com/fezcode/gobake"
)

func main() {
	bake := gobake.NewEngine()

	// Load metadata
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		fmt.Println("Error loading recipe.piml:", err)
		return
	}

	// --- Tasks ---

	bake.Task("build", "Compiles the application", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s...", bake.Info.Name, bake.Info.Version)
		// Example: Build for the current OS
		// return ctx.Run("go", "build", "-o", "bin/app", ".")
		return nil
	})

	bake.Task("test", "Runs project tests", func(ctx *gobake.Context) error {
		ctx.Log("Running tests...")
		return ctx.Run("go", "test", "./...")
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		ctx.Log("Cleaning up...")
		// Using shell command for cross-platform delete is tricky in raw shell,
		// but since we are in Go, we can use os.RemoveAll!
		// But ctx.Run uses shell.
		return ctx.Run("go", "clean")
	})

	// Execute the task passed in CLI
	bake.Execute()
}
`

	if err := os.WriteFile("recipe.piml", []byte(pimlContent), 0644); err != nil {
		fmt.Printf("Error creating recipe.piml: %v\n", err)
		return
	}
	fmt.Println("Created recipe.piml")

	if err := os.WriteFile("Recipe.go", []byte(recipeGoContent), 0644); err != nil {
		fmt.Printf("Error creating Recipe.go: %v\n", err)
		return
	}
	fmt.Println("Created Recipe.go")

	fmt.Println("\nSuccess! initialized gobake project.")
	fmt.Println("Run 'gobake build' to start.")
}
