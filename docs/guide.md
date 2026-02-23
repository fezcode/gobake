# gobake User Guide

**gobake** is a Go-native build orchestrator. It allows you to write your build logic in Go, benefiting from type safety, autocomplete, and the full power of the Go standard library, without needing a separate build language like Make or CMake.

## 1. Installation

```bash
go install github.com/fezcode/gobake/cmd/gobake@latest
```

## 2. Project Structure

A typical gobake project looks like this:

```text
my-project/
├── go.mod
├── main.go
├── recipe.piml   <-- Metadata (Version, Tools, Name)
└── Recipe.go     <-- Build Logic
```

### The `recipe.piml` File
This file acts as the "Manifest" for your project. It stores metadata that doesn't belong in code.

```piml
(name) my-app
(version) 1.0.0
(description) A sample application
(authors)
  > John Doe
(license) MIT
(tools)
  > golang.org/x/tools/cmd/stringer@latest
  > github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### The `Recipe.go` File
This is where you define your tasks.

```go
//go:build gobake
package bake_recipe

import "github.com/fezcode/gobake"

func Run(bake *gobake.Engine) error {
    if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
        return err
    }

    bake.Task("build", "Build the binary", func(ctx *gobake.Context) error {
        return ctx.Run("go", "build", "-o", "bin/app", ".")
    })
    
    return nil
}
```

## 3. How It Works (The "Magic")

You might wonder about `//go:build gobake` and `package bake_recipe`. Here is what happens under the hood when you run `gobake build`:

1.  **Isolation**: The `//go:build gobake` build tag tells the standard Go compiler (when you run `go build` on your project) to **ignore** this file. This prevents your build logic from being compiled into your actual application binary.
2.  **Package Separation**: The file uses `package bake_recipe` to keep it in a separate namespace from your project's `main` package.
3.  **Generation & Execution**:
    *   `gobake` reads your `Recipe.go`.
    *   It creates a hidden temporary directory `.gobake/`.
    *   It copies your code there, changes the package to `main`, and removes the build tag.
    *   It generates a small runner file that initializes the `gobake.Engine`, calls your `Run()` function, and executes the requested task.
    *   Finally, it executes this generated program using `go run`.

This process allows you to write "script-like" Go code that is fully compiled and type-checked on the fly.

## 4. CLI Reference

### Project Initialization
*   **`gobake init`**: Scaffolds a new project.
    *   Creates `recipe.piml` and `Recipe.go`.
    *   Runs `go mod init` (if missing) and `go mod tidy`.
    *   Adds `github.com/fezcode/gobake` as a dependency.
*   **`gobake template <git-url>`**: Clones a git repo and initializes it as a gobake project.

### Dependency Management
*   **`gobake add-tool <url>`**: Adds a Go tool to `recipe.piml` (e.g., `gobake add-tool github.com/swaggo/swag/cmd/swag@latest`).
*   **`gobake remove-tool <url>`**: Removes a tool.
*   **`gobake add-dep <url>`**: Adds a library dependency (`go get`).
*   **`gobake remove-dep <url>`**: Removes a library dependency.

### Versioning
*   **`gobake bump [patch|minor|major]`**: Automatically increments the version in `recipe.piml`.
    *   `patch`: 1.0.0 -> 1.0.1
    *   `minor`: 1.0.1 -> 1.1.0
    *   `major`: 1.1.0 -> 2.0.0

### General
*   **`gobake help`**: Lists all available commands AND the tasks defined in your `Recipe.go`.
*   **`gobake version`**: Shows the gobake CLI version.
