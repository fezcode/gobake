package gobake

import (
	"fmt"
	"os"
	"os/exec"
)

// Task represents a build action.
type Task struct {
	Name        string
	Description string
	Action      func(ctx *Context) error
}

// Context provides utilities for tasks.
type Context struct {
	Engine *Engine
	Args   []string
}

// Engine manages tasks and execution.
type Engine struct {
	Tasks map[string]*Task
	Info  *RecipeInfo
}

// RecipeInfo holds metadata from recipe.piml.
type RecipeInfo struct {
	Name        string
	Version     string
	Description string
	Authors     []string
	License     string
	Repository  string
	Homepage    string
	Keywords    []string
	Tools       []string
}

func New() *Engine {
	return NewEngine()
}

func NewEngine() *Engine {
	return &Engine{
		Tasks: make(map[string]*Task),
	}
}

// InstallTools installs Go tools listed in recipe.piml.
func (ctx *Context) InstallTools() error {
	if ctx.Engine.Info == nil || len(ctx.Engine.Info.Tools) == 0 {
		ctx.Log("No tools defined in recipe.piml")
		return nil
	}

	for _, tool := range ctx.Engine.Info.Tools {
		ctx.Log("Installing tool: %s", tool)
		if err := ctx.Run("go", "install", tool); err != nil {
			return err
		}
	}
	return nil
}

// Task registers a new task.
func (e *Engine) Task(name, description string, action func(ctx *Context) error) {
	if ctx.Engine.Info == nil || len(ctx.Engine.Info.Tools) == 0 {
		ctx.Log("No tools defined in recipe.piml")
		return nil
	}

	for _, tool := range ctx.Engine.Info.Tools {
		ctx.Log("Installing tool: %s", tool)
		if err := ctx.Run("go", "install", tool); err != nil {
			return err
		}
	}
	return nil
}

// BakeBinary cross-compiles a Go binary.
func (ctx *Context) BakeBinary(osName, arch, output string, flags ...string) error {
	ctx.Log("Baking binary for %s/%s -> %s", osName, arch, output)
	
	args := []string{"build"}
	args = append(args, flags...)
	args = append(args, "-o", output, ".")

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), 
		"GOOS="+osName,
		"GOARCH="+arch,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Log prints a formatted message to stdout.
func (ctx *Context) Log(format string, a ...interface{}) {
	fmt.Printf("[gobake] "+format+"\n", a...)
}

// Run executes a shell command and waits for it to finish.
func (ctx *Context) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Execute runs the engine based on command line arguments.
func (e *Engine) Execute() {
	if len(os.Args) < 2 {
		e.PrintHelp()
		return
	}

	taskName := os.Args[1]
	task, ok := e.Tasks[taskName]
	if !ok {
		fmt.Printf("Unknown task: %s\n", taskName)
		e.PrintHelp()
		os.Exit(1)
	}

	ctx := &Context{
		Engine: e,
		Args:   os.Args[2:],
	}

	if err := task.Action(ctx); err != nil {
		fmt.Printf("Task '%s' failed: %v\n", taskName, err)
		os.Exit(1)
	}
}

func (e *Engine) PrintHelp() {
	fmt.Println("Usage: gobake <task> [args]")
	fmt.Println("\nAvailable tasks:")
	for name, task := range e.Tasks {
		fmt.Printf("  %-15s %s\n", name, task.Description)
	}
}
