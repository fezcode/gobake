# gobake API Reference

This document covers the core types and functions available in `gobake`.

## Core Types

*   `gobake.Engine`: The central controller for defining tasks and loading project metadata.
*   `gobake.Context`: A helper passed to each task, providing execution and utility methods.
*   `gobake.Task`: Represents a single build step.
*   `gobake.RecipeInfo`: Struct mapping to `recipe.piml`.

## 1. Engine API

### `func (e *Engine) Task(name, description string, action func(ctx *Context) error)`

Registers a new task.
*   **name**: The name of the task (e.g., "build", "test"). Must be unique.
*   **description**: A short description shown in `gobake help`.
*   **action**: The function to execute. Returns `error`.

```go
bake.Task("clean", "Cleans the build artifacts", func(ctx *gobake.Context) error {
    return ctx.Remove("bin/")
})
```

### `func (e *Engine) TaskWithDeps(name, description string, deps []string, action func(ctx *Context) error)`

Registers a task that depends on other tasks. Dependencies run **before** the main action.
*   **deps**: A slice of task names (e.g., `[]string{"test", "lint"}`).

```go
bake.TaskWithDeps("build", "Build app", []string{"clean"}, func(ctx *gobake.Context) error {
    return ctx.BakeBinary("linux", "amd64", "bin/app")
})
```

### `func (e *Engine) LoadRecipeInfo(path string) error`

Loads project metadata from `recipe.piml` into `e.Info`.
*   **path**: Usually `"recipe.piml"`.

```go
if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
    return err
}
fmt.Println("Project Name:", bake.Info.Name)
```

### `func (e *Engine) SaveRecipeInfo(path string) error`

Saves the current `e.Info` back to `recipe.piml`. Useful for custom scripts that modify version or metadata.

## 2. Context API

### Execution

#### `func (ctx *Context) Run(name string, args ...string) error`
Executes a shell command. It inherits stdout/stderr and environment variables.
*   **name**: The command (e.g., "go", "npm", "docker").
*   **args**: Command arguments.

```go
ctx.Run("go", "test", "./...")
```

#### `func (ctx *Context) RunIn(dir, name string, args ...string) error`
Like `Run`, but executes the command in the given working directory. An empty `dir` falls back to the current working directory.

```go
ctx.RunIn("./service", "go", "build", "./...")
```

#### `func (ctx *Context) RunOutput(name string, args ...string) (string, error)`
Executes a command and returns its captured stdout (with trailing newlines trimmed). Stderr is still streamed to the terminal so failures stay visible.

```go
sha, err := ctx.RunOutput("git", "rev-parse", "HEAD")
if err != nil {
    return err
}
ctx.Log("commit: %s", sha)
```

#### `func (ctx *Context) RunInOutput(dir, name string, args ...string) (string, error)`
`RunOutput` with an explicit working directory.

```go
branch, _ := ctx.RunInOutput("./vendor/lib", "git", "rev-parse", "--abbrev-ref", "HEAD")
```

#### `func (ctx *Context) BakeBinary(osName, arch, output string, flags ...string) error`
A helper for cross-compiling Go binaries. Sets `GOOS` and `GOARCH` automatically.
*   **osName**: Target OS (e.g., "linux", "windows").
*   **arch**: Target Architecture (e.g., "amd64", "arm64").
*   **output**: Output file path.
*   **flags**: Additional `go build` flags (e.g., `"-ldflags", "-s -w"`).

```go
ctx.BakeBinary("windows", "amd64", "bin/app.exe", "-v")
```

#### `func (ctx *Context) InstallTools() error`
Installs all Go tools listed in `recipe.piml` -> `(tools)`.

```go
bake.Task("setup", "Install dev tools", func(ctx *gobake.Context) error {
    return ctx.InstallTools()
})
```

### File System

#### `func (ctx *Context) Mkdir(path string) error`
Creates a directory and any necessary parents (like `mkdir -p`).

#### `func (ctx *Context) Remove(path string) error`
Removes a file or directory recursively (like `rm -rf`).

#### `func (ctx *Context) Copy(src, dst string) error`
Copies a single file from `src` to `dst`.

### Utilities

#### `func (ctx *Context) Log(format string, a ...interface{})`
Prints a formatted message to stdout, prefixed with `[gobake]`.

#### `func (ctx *Context) SetEnv(key, value string)`
Sets an environment variable for subsequent `Run` or `BakeBinary` calls within the same task context.

```go
ctx.SetEnv("CGO_ENABLED", "0")
ctx.BakeBinary("linux", "amd64", "bin/app")
```
