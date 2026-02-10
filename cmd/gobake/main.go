package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Handle "init" command
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runInit()
		return
	}

	// Handle "bump" command
	if len(os.Args) > 1 && os.Args[1] == "bump" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: gobake bump [patch|minor|major]")
			return
		}
		runBump(os.Args[2])
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
    - golang
(tools)
    - github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

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

	bake.Task("setup", "Installs required tools", func(ctx *gobake.Context) error {
		return ctx.InstallTools()
	})

	bake.Task("build", "Compiles the application", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s...", bake.Info.Name, bake.Info.Version)
		return ctx.BakeBinary("linux", "amd64", "bin/app-linux")
	})

	bake.Task("test", "Runs project tests", func(ctx *gobake.Context) error {
		ctx.Log("Running tests...")
		return ctx.Run("go", "test", "./...")
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		ctx.Log("Cleaning up...")
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

func runBump(part string) {
	data, err := os.ReadFile("recipe.piml")
	if err != nil {
		fmt.Printf("Error reading recipe.piml: %v\n", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "(version)") {
			currentVersion := strings.TrimSpace(strings.TrimPrefix(line, "(version)"))
			newVersion, err := incrementVersion(currentVersion, part)
			if err != nil {
				fmt.Printf("Error incrementing version: %v\n", err)
				return
			}
			lines[i] = fmt.Sprintf("(version) %s", newVersion)
			fmt.Printf("Bumped version: %s -> %s\n", currentVersion, newVersion)
			found = true
			break
		}
	}

	if !found {
		fmt.Println("Error: (version) tag not found in recipe.piml")
		return
	}

	if err := os.WriteFile("recipe.piml", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		fmt.Printf("Error writing recipe.piml: %v\n", err)
		return
	}
}

func incrementVersion(version, part string) (string, error) {
	var major, minor, patch int
	_, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		// Try 2-part version just in case
		_, err = fmt.Sscanf(version, "%d.%d", &major, &minor)
		if err != nil {
			return "", fmt.Errorf("invalid version format (expected X.Y.Z): %s", version)
		}
	}

	switch strings.ToLower(part) {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("invalid bump part: %s (use major, minor, or patch)", part)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}
