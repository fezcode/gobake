# gobake Examples

## 1. Cross-Compilation for Multiple Platforms

This recipe builds your application for Windows, Linux, and macOS.

```go
package main

import (
    "fmt"
    "github.com/fezcode/gobake"
)

func main() {
    bake := gobake.NewEngine()
    bake.LoadRecipeInfo("recipe.piml")

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
            
            err := ctx.BakeBinary(p.OS, p.Arch, output, "-ldflags", "-s -w")
            if err != nil {
                return err
            }
        }
        return nil
    })

    bake.Execute()
}
```

## 2. Using External Tools (e.g., Code Generation)

This recipe installs `stringer` and uses it to generate code before building.

**recipe.piml:**
```piml
(name) my-app
(tools)
    > golang.org/x/tools/cmd/stringer@latest
```

**Recipe.go:**
```go
package main

import "github.com/fezcode/gobake"

func main() {
    bake := gobake.NewEngine()
    bake.LoadRecipeInfo("recipe.piml")

    bake.Task("generate", "Generates code", func(ctx *gobake.Context) error {
        // Ensure tools are installed first
        if err := ctx.InstallTools(); err != nil {
            return err
        }
        ctx.Log("Running stringer...")
        return ctx.Run("go", "generate", "./...")
    })

    bake.Task("build", "Builds app", func(ctx *gobake.Context) error {
        if err := ctx.Run("gobake", "generate"); err != nil {
            return err
        }
        return ctx.Run("go", "build", ".")
    })

    bake.Execute()
}
```

## 3. Injecting Version with ldflags

This recipe automatically injects the version from `recipe.piml` into your binary's `main.Version` variable.

**main.go:**
```go
package main

import "fmt"

var Version = "dev" // Default value

func main() {
    fmt.Println("Version:", Version)
}
```

**Recipe.go:**
```go
package main

import (
    "fmt"
    "github.com/fezcode/gobake"
)

func main() {
    bake := gobake.NewEngine()
    bake.LoadRecipeInfo("recipe.piml")

    bake.Task("build", "Build with version", func(ctx *gobake.Context) error {
        // -X main.Version=1.2.3
        ldflags := fmt.Sprintf("-X main.Version=%s", bake.Info.Version)
        
        ctx.Log("Building version %s...", bake.Info.Version)
        return ctx.Run("go", "build", "-ldflags", ldflags, "-o", "bin/app")
    })

    bake.Execute()
}
```
