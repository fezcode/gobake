# gobake User Guide

## Core Concepts

`gobake` operates on a simple principle: **Your build system should be as robust as your application.**

Instead of relying on fragile shell scripts or complex Makefiles, `gobake` uses:
1.  **Engine:** The core Go library that manages tasks.
2.  **Context:** A helper object passed to every task, providing access to the shell, logging, and metadata.
3.  **Recipe:** The `Recipe.go` file that ties it all together.

## CLI Command Reference

### Project Management

*   **`gobake init`**: Initialize a new project in the current directory.
*   **`gobake version`**: Show the current version of gobake.
*   **`gobake help`**: Show the list of available commands and project tasks.
*   **`gobake template <git-url>`**: Clone a repository and initialize it.

### Dependency & Tool Management

*   **`gobake add-tool <url>`**: Adds a tool to the `(tools)` list in `recipe.piml`.
    *   Example: `gobake add-tool github.com/vektra/mockery/v2@latest`
*   **`gobake remove-tool <url>`**: Removes a tool from `recipe.piml`.
*   **`gobake add-dep <url>`**: Adds a library dependency to `go.mod` (wrapper for `go get`).
    *   Example: `gobake add-dep github.com/gin-gonic/gin`
*   **`gobake remove-dep <url>`**: Removes a library dependency.

### Versioning

*   **`gobake bump [patch|minor|major]`**: Increments the project version in `recipe.piml`.

## Metadata (`recipe.piml`)

The `recipe.piml` file is the single source of truth for your project.

### Example

```piml
(name) my-project
(version) 1.2.3
(authors)
    > Alice <alice@example.com>
    > Bob <bob@example.com>
(tools)
    > github.com/swaggo/swag/cmd/swag@latest
```

### Fields

*   `name`: The name of your project.
*   `version`: Semantic version (e.g., `1.0.0`).
*   `description`: A short summary.
*   `authors`: List of contributors.
*   `license`: Project license (e.g., `MIT`, `Apache-2.0`).
*   `repository`: Git repository URL.
*   `keywords`: Search tags.
*   `tools`: List of External CLI tools required for development (e.g., linters, generators).

### Dependency Management vs. Tools

It is important to distinguish between **Project Dependencies** and **Build Tools**.

1.  **Project Dependencies (`go.mod`):**
    *   Libraries your code imports (e.g., `gin`, `cobra`).
    *   Managed by standard Go commands: `go get`, `go mod tidy`.
    *   Do **not** list these in `recipe.piml`.

2.  **Build Tools (`recipe.piml`):**
    *   Executables used *during* the build process (e.g., `golangci-lint`, `stringer`, `swag`).
    *   These are installed via `go install` into your system or path.
    *   `gobake` automates installing these using `ctx.InstallTools()`.

### Accessing Metadata in Code

You can access these fields in your `Recipe.go` via `bake.Info`.

```go
fmt.Println("Project:", bake.Info.Name)
fmt.Println("Version:", bake.Info.Version)
```

## Task Management

Tasks are defined using `bake.Task(name, description, function)`.

### Task Dependencies

You can define tasks that depend on other tasks using `bake.TaskWithDeps`. Dependencies are executed sequentially before the main task.

```go
bake.Task("test", "Run tests", func(ctx *gobake.Context) error { ... })
bake.TaskWithDeps("build", "Build app", []string{"test"}, func(ctx *gobake.Context) error { ... })
```

`gobake` automatically detects circular dependencies and ensures each task runs only once per execution.

### Reserved Task Names

The following names are reserved for `gobake` CLI commands and cannot be used as task names:
*   `init`
*   `version`
*   `help`
*   `bump`
*   `template`
*   `add-tool`
*   `remove-tool`
*   `add-dep`
*   `remove-dep`

### Context Methods

The `ctx` object is powerful. Here are its key methods:

*   **`ctx.Run(cmd, args...)`**: Executes a shell command. It streams stdout/stderr to your terminal.
*   **`ctx.Log(format, args...)`**: Prints a formatted log message with the `[gobake]` prefix.
*   **`ctx.InstallTools()`**: Iterates through the `tools` list in `recipe.piml` and runs `go install` for each.
*   **`ctx.BakeBinary(os, arch, output, flags...)`**: A helper for cross-compilation. It sets `GOOS` and `GOARCH` automatically.
*   **`ctx.Mkdir(path)`**: Creates a directory and all its parents.
*   **`ctx.Remove(path)`**: Removes a file or directory recursively.
*   **`ctx.Copy(src, dst)`**: Copies a file.
*   **`ctx.SetEnv(key, value)`**: Sets an environment variable **only** for the current task and commands executed within it via `ctx.Run` or `ctx.BakeBinary`.

## Workflow Tips

*   **Chaining Tasks:** You can run one task from another.
    ```go
    bake.Task("all", "Builds everything", func(ctx *gobake.Context) error {
        if err := ctx.Run("gobake", "test"); err != nil { return err }
        return ctx.Run("gobake", "build")
    })
    ```
*   **CI/CD:** Since `gobake` is just Go, it runs perfectly in GitHub Actions or GitLab CI. Just ensure Go is installed.
*   **Injecting Version:** You can inject the version from `recipe.piml` into your binary using `-ldflags`.
    ```go
    bake.Task("build", "Build app", func(ctx *gobake.Context) error {
        ldflags := fmt.Sprintf("-X main.Version=%s", bake.Info.Version)
        return ctx.BakeBinary("linux", "amd64", "bin/app", "-ldflags", ldflags)
    })
    ```
