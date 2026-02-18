package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fezcode/gobake"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	defaultRecipePiml, _ = templatesFS.ReadFile("templates/recipe.piml")
	defaultRecipeGo, _   = templatesFS.ReadFile("templates/Recipe.go.tmpl")
)

func main() {
	// Handle "version" command
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "-v") {
		fmt.Printf("gobake version %s\n", gobake.Version)
		return
	}

	// Handle "help" command
	if len(os.Args) > 1 && (os.Args[1] == "help" || os.Args[1] == "--help") {
		printCliHelp()
		return
	}

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

	// Handle "template" command
	if len(os.Args) > 1 && os.Args[1] == "template" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: gobake template <git-repo-url>")
			return
		}
		runTemplate(os.Args[2])
		return
	}

	// Handle "add-tool" command
	if len(os.Args) > 1 && os.Args[1] == "add-tool" {
		if len(os.Args) < 2 {
			fmt.Println("Usage: gobake add-tool <tool-package-url>")
			return
		}
		runAddTool(os.Args[2])
		return
	}

	// Handle "add-dep" command
	if len(os.Args) > 1 && os.Args[1] == "add-dep" {
		if len(os.Args) < 2 {
			fmt.Println("Usage: gobake add-dep <package-url>")
			return
		}
		runAddDep(os.Args[2])
		return
	}

	// Handle "remove-tool" command
	if len(os.Args) > 1 && os.Args[1] == "remove-tool" {
		if len(os.Args) < 2 {
			fmt.Println("Usage: gobake remove-tool <tool-package-url>")
			return
		}
		runRemoveTool(os.Args[2])
		return
	}

	// Handle "remove-dep" command
	if len(os.Args) > 1 && os.Args[1] == "remove-dep" {
		if len(os.Args) < 2 {
			fmt.Println("Usage: gobake remove-dep <package-url>")
			return
		}
		runRemoveDep(os.Args[2])
		return
	}

	// Check for Recipe.go
	if _, err := os.Stat("Recipe.go"); err == nil {
		if err := runRecipe(os.Args[1:]); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		return
	}

	// If no Recipe.go, check for recipe.piml to give a hint
	if _, err := os.Stat("recipe.piml"); err == nil {
		fmt.Println("Found recipe.piml but no Recipe.go.")
		fmt.Println("Please create a Recipe.go file to define your build tasks.")
		fmt.Println("\nExample Recipe.go:")
		fmt.Printf("%s\n", `//go:build gobake
package bake_recipe

import (
	"fmt"
	"github.com/fezcode/gobake"
)

func Run(bake *gobake.Engine) error {
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		return fmt.Errorf("error loading recipe.piml: %v", err)
	}
	
	bake.Task("build", "Build the project", func(ctx *gobake.Context) error {
		return ctx.Run("go", "build", "-o", "bin/app")
	})
	
	return nil
}`)
		return
	}

	fmt.Println("gobake: No Recipe.go or recipe.piml found in the current directory.")
	fmt.Println("Run 'gobake init' to create a new project configuration.")
	fmt.Println("Visit https://github.com/fezcode/gobake for more information.")
}

func runRecipe(programArgs []string) error {
	// 1. Read Recipe.go
	content, err := os.ReadFile("Recipe.go")
	if err != nil {
		return fmt.Errorf("error reading Recipe.go: %v", err)
	}

	// Create hidden directory
	tmpDir := ".gobake"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. Prepare gobake_recipe_gen.go
	// Replace 'package bake_recipe' with 'package main'
	// Also remove '//go:build ignore' or '//go:build gobake' to ensure it's included in the run
	sContent := string(content)
	sContent = strings.Replace(sContent, "package bake_recipe", "package main", 1)
	sContent = strings.Replace(sContent, "//go:build ignore", "", 1)
	sContent = strings.Replace(sContent, "//go:build gobake", "", 1)
	
	recipeGenFile := filepath.Join(tmpDir, "gobake_recipe_gen.go")
	if err := os.WriteFile(recipeGenFile, []byte(sContent), 0644); err != nil {
		return fmt.Errorf("error creating temporary recipe file: %v", err)
	}

	// 3. Create gobake_runner_gen.go
	runnerContent := `package main

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
`
	runnerGenFile := filepath.Join(tmpDir, "gobake_runner_gen.go")
	if err := os.WriteFile(runnerGenFile, []byte(runnerContent), 0644); err != nil {
		return fmt.Errorf("error creating runner file: %v", err)
	}

	// 4. Run go run .gobake/gobake_recipe_gen.go .gobake/gobake_runner_gen.go [args]
	args := append([]string{"run", recipeGenFile, runnerGenFile}, programArgs...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
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

	// 1. Handle go.mod if it doesn't exist
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		dir, _ := os.Getwd()
		modName := filepath.Base(dir)
		fmt.Printf("go.mod not found. Running 'go mod init %s'...\n", modName)
		cmd := exec.Command("go", "mod", "init", modName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: 'go mod init' failed: %v\n", err)
		}
	}

	// 2. Write embedded templates
	if err := os.WriteFile("recipe.piml", defaultRecipePiml, 0644); err != nil {
		fmt.Printf("Error creating recipe.piml: %v\n", err)
		return
	}
	fmt.Println("Created recipe.piml")

	if err := os.WriteFile("Recipe.go", defaultRecipeGo, 0644); err != nil {
		fmt.Printf("Error creating Recipe.go: %v\n", err)
		return
	}
	fmt.Println("Created Recipe.go")

	// 3. Get gobake library
	fmt.Println("Getting gobake library...")
	getCmd := exec.Command("go", "get", "github.com/fezcode/gobake")
	getCmd.Stdout = os.Stdout
	getCmd.Stderr = os.Stderr
	if err := getCmd.Run(); err != nil {
		fmt.Printf("Warning: 'go get github.com/fezcode/gobake' failed: %v\n", err)
	}

	// 4. Run go mod tidy to resolve and keep dependencies
	fmt.Println("Running 'go mod tidy' to resolve dependencies...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		fmt.Printf("Warning: 'go mod tidy' failed: %v\n", err)
	}

	fmt.Println("\nSuccess! initialized gobake project.")
	fmt.Println("Run 'gobake build' to start.")
}

func runBump(part string) {
	bake := gobake.NewEngine()
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		fmt.Printf("Error reading recipe.piml: %v\n", err)
		return
	}

	currentVersion := bake.Info.Version
	newVersion, err := incrementVersion(currentVersion, part)
	if err != nil {
		fmt.Printf("Error incrementing version: %v\n", err)
		return
	}

	bake.Info.Version = newVersion
	if err := bake.SaveRecipeInfo("recipe.piml"); err != nil {
		fmt.Printf("Error writing recipe.piml: %v\n", err)
		return
	}

	fmt.Printf("Bumped version: %s -> %s\n", currentVersion, newVersion)
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

func runTemplate(repoUrl string) {
	// 1. Determine directory name from URL
	// e.g. https://github.com/user/project.git -> project
	base := filepath.Base(repoUrl)
	dirName := strings.TrimSuffix(base, ".git")
	if dirName == "" || dirName == "." || dirName == "/" {
		dirName = "new-project"
	}

	// 2. Git Clone
	fmt.Printf("Cloning %s into %s...\n", repoUrl, dirName)
	cmd := exec.Command("git", "clone", repoUrl, dirName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		return
	}

	// 3. Check and create recipe.piml
	pimlPath := filepath.Join(dirName, "recipe.piml")
	if _, err := os.Stat(pimlPath); os.IsNotExist(err) {
		fmt.Println("recipe.piml not found. Creating default...")

		if err := os.WriteFile(pimlPath, defaultRecipePiml, 0644); err != nil {
			fmt.Printf("Error creating recipe.piml: %v\n", err)
		} else {
			fmt.Println("Created recipe.piml")
		}
	} else {
		fmt.Println("recipe.piml already exists.")
	}

	// 4. Check and create Recipe.go
	recipeGoPath := filepath.Join(dirName, "Recipe.go")
	if _, err := os.Stat(recipeGoPath); os.IsNotExist(err) {
		fmt.Println("Recipe.go not found. Creating default...")

		if err := os.WriteFile(recipeGoPath, defaultRecipeGo, 0644); err != nil {
			fmt.Printf("Error creating Recipe.go: %v\n", err)
		} else {
			fmt.Println("Created Recipe.go")
		}
	} else {
		fmt.Println("Recipe.go already exists.")
	}

	fmt.Printf("\nProject ready in directory: %s\n", dirName)
	fmt.Printf("cd %s\n", dirName)
	fmt.Println("gobake build")
}

func runAddDep(pkg string) {
	fmt.Printf("Adding dependency: %s...\n", pkg)
	cmd := exec.Command("go", "get", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error adding dependency: %v\n", err)
	}
}

func runAddTool(tool string) {
	bake := gobake.NewEngine()
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		fmt.Printf("Error reading recipe.piml: %v\n", err)
		return
	}

	// Check if already exists
	for _, t := range bake.Info.Tools {
		if t == tool {
			fmt.Printf("Tool %s already exists in recipe.piml\n", tool)
			return
		}
	}

	bake.Info.Tools = append(bake.Info.Tools, tool)

	if err := bake.SaveRecipeInfo("recipe.piml"); err != nil {
		fmt.Printf("Error updating recipe.piml: %v\n", err)
		return
	}
	fmt.Printf("Added tool %s to recipe.piml\n", tool)
}

func runRemoveDep(pkg string) {
	fmt.Printf("Removing dependency: %s...\n", pkg)
	// 'go get package@none' removes it from go.mod
	cmd := exec.Command("go", "get", pkg+"@none")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error removing dependency: %v\n", err)
	}
}

func runRemoveTool(tool string) {
	bake := gobake.NewEngine()
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		fmt.Printf("Error reading recipe.piml: %v\n", err)
		return
	}

	var newTools []string
	found := false
	for _, t := range bake.Info.Tools {
		if t == tool {
			found = true
			continue
		}
		newTools = append(newTools, t)
	}

	if !found {
		fmt.Printf("Tool '%s' not found in recipe.piml\n", tool)
		return
	}

	bake.Info.Tools = newTools

	if err := bake.SaveRecipeInfo("recipe.piml"); err != nil {
		fmt.Printf("Error updating recipe.piml: %v\n", err)
		return
	}
	fmt.Printf("Removed tool %s from recipe.piml\n", tool)
}

func printCliHelp() {
	fmt.Println("gobake - Go-native build orchestrator")
	fmt.Printf("Version: %s\n", gobake.Version)
	fmt.Println("\nUsage: gobake <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  init          Initialize a new project")
	fmt.Println("  version       Show gobake version")
	fmt.Println("  bump          Bump project version (patch|minor|major)")
	fmt.Println("  template      Init from a git template")
	fmt.Println("  add-tool      Add a tool dependency")
	fmt.Println("  remove-tool   Remove a tool dependency")
	fmt.Println("  add-dep       Add a library dependency")
	fmt.Println("  remove-dep    Remove a library dependency")
	fmt.Println("  help          Show this help")

	if _, err := os.Stat("Recipe.go"); err == nil {
		fmt.Println("\n--- Project Tasks ---")
		// We run the recipe without arguments to trigger the default behavior (showing help/tasks if no task specified)
		// But runRecipe expects args to pass to the binary.
		// If we pass empty args, the runner's main() calls bake.Execute().
		// gobake.Execute() checks os.Args. If len(os.Args) < 2, it prints help.
		// The runner is: "go run gen_recipe.go gen_runner.go [args...]"
		// So os.Args[0] is the binary, os.Args[1] is the first arg.
		// We want to simulate "gobake" (no args) behavior for the runner.
		// runRecipe appends passed args to the go run command.
		// If we pass [], command is `go run ...`.
		// Inside the runner, os.Args will be [binary_path].
		// gobake.Execute() checks len(os.Args) < 2. So it will print help.
		runRecipe([]string{})
	}
}
