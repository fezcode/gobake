# gobake Examples

Here are several practical examples of how to use `gobake` for different scenarios.

## 1. Basic Build and Test

A standard Go project setup.

**recipe.piml:**
```piml
(name) my-app
(version) 1.0.0
(tools)
    > golang.org/x/tools/cmd/stringer@latest
```

**Recipe.go:**
```go
//go:build gobake
package bake_recipe

import "github.com/fezcode/gobake"

func Run(bake *gobake.Engine) error {
    // 1. Load project info
    if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
        return err
    }

    // 2. Setup tools
    bake.Task("tools", "Install dev dependencies", func(ctx *gobake.Context) error {
        return ctx.InstallTools()
    })

    // 3. Test
    bake.TaskWithDeps("test", "Run tests", []string{"tools"}, func(ctx *gobake.Context) error {
        return ctx.Run("go", "test", "./...")
    })

    // 4. Build
    bake.TaskWithDeps("build", "Build binary", []string{"test"}, func(ctx *gobake.Context) error {
        return ctx.BakeBinary("linux", "amd64", "bin/app")
    })

    return nil
}
```

## 2. Cross-Compilation Matrix

Build for multiple operating systems and architectures in one go.

```go
bake.Task("release", "Build for all platforms", func(ctx *gobake.Context) error {
    platforms := []struct {
        OS   string
        Arch string
        Ext  string
    }{
        {"linux", "amd64", ""},
        {"windows", "amd64", ".exe"},
        {"darwin", "arm64", ""},
    }

    for _, p := range platforms {
        output := fmt.Sprintf("dist/%s-%s-%s%s", 
            bake.Info.Name, p.OS, p.Arch, p.Ext)
        
        ctx.Log("Building for %s/%s...", p.OS, p.Arch)
        // -ldflags "-s -w" strips debug info for smaller binaries
        err := ctx.BakeBinary(p.OS, p.Arch, output, "-ldflags", "-s -w")
        if err != nil {
            return err
        }
    }
    return nil
})
```

## 3. Version Injection (ldflags)

Automatically inject the version from `recipe.piml` into your application.

**main.go:**
```go
package main
var Version = "dev"
func main() { println("Version:", Version) }
```

**Recipe.go:**
```go
bake.Task("build", "Build with version", func(ctx *gobake.Context) error {
    // Inject bake.Info.Version into main.Version
    ldflags := fmt.Sprintf("-X main.Version=%s", bake.Info.Version)
    
    return ctx.Run("go", "build", "-ldflags", ldflags, "-o", "bin/app")
})
```

## 4. Code Generation & Tools

Using `go generate` or external tools before building.

```go
bake.Task("generate", "Run go generate", func(ctx *gobake.Context) error {
    // Ensure stringer, mockgen, etc. are installed
    if err := ctx.InstallTools(); err != nil {
        return err
    }
    return ctx.Run("go", "generate", "./...")
})
```

## 5. Clean & Coverage

Using context helpers for file operations.

```go
bake.Task("clean", "Remove artifacts", func(ctx *gobake.Context) error {
    if err := ctx.Remove("bin/"); err != nil {
        return err
    }
    return ctx.Remove("coverage.out")
})

bake.Task("coverage", "Run tests with coverage", func(ctx *gobake.Context) error {
    if err := ctx.Run("go", "test", "-coverprofile=coverage.out", "./..."); err != nil {
        return err
    }
    return ctx.Run("go", "tool", "cover", "-html=coverage.out")
})
```

## 6. Capturing Command Output

Use `RunOutput` when you need a command's stdout — for example, embedding the current git commit into a binary.

```go
bake.Task("build", "Build with git SHA", func(ctx *gobake.Context) error {
    sha, err := ctx.RunOutput("git", "rev-parse", "--short", "HEAD")
    if err != nil {
        return err
    }
    ldflags := fmt.Sprintf("-X main.Version=%s -X main.Commit=%s",
        bake.Info.Version, sha)
    return ctx.Run("go", "build", "-ldflags", ldflags, "-o", "bin/app")
})
```

## 7. Running Commands in Subdirectories

`RunIn` (and `RunInOutput`) execute a command in a specific working directory without mutating global state — safe to call from concurrent tasks.

```go
bake.Task("vendor-build", "Build a vendored sub-project", func(ctx *gobake.Context) error {
    return ctx.RunIn("./third_party/lib", "go", "build", "./...")
})
```

## 8. CI/CD Integration

Since `gobake` is just a Go binary, you can run it easily in CI environments.

**GitHub Actions (.github/workflows/build.yml):**
```yaml
name: Build
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.20'
      
      # Install gobake
      - run: go install github.com/fezcode/gobake/cmd/gobake@latest
      
      # Run tasks
      - run: gobake test
      - run: gobake build
```
