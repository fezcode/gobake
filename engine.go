package gobake

import (
	"fmt"
	"os"
	"os/exec"
)

const Version = "0.1.0"

// Task represents a build action.
type Task struct {
	Name        string
	Description string
	Action      func(ctx *Context) error
	DependsOn   []string
}

// Context provides utilities for tasks.
type Context struct {
	Engine *Engine
	Args   []string
}

// Engine manages tasks and execution.
type Engine struct {
	Tasks         map[string]*Task
	Info          *RecipeInfo
	executedTasks map[string]bool
}

// RecipeInfo holds metadata from recipe.piml.
type RecipeInfo struct {
	Name        string   `piml:"name"`
	Version     string   `piml:"version,omitempty"`
	Description string   `piml:"description,omitempty"`
	Authors     []string `piml:"authors,omitempty"`
	License     string   `piml:"license,omitempty"`
	Repository  string   `piml:"repository,omitempty"`
	Homepage    string   `piml:"homepage,omitempty"`
	Keywords    []string `piml:"keywords,omitempty"`
	Tools       []string `piml:"tools,omitempty"`
}

func New() *Engine {
	return NewEngine()
}

func NewEngine() *Engine {
	return &Engine{
		Tasks:         make(map[string]*Task),
		executedTasks: make(map[string]bool),
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
	e.TaskWithDeps(name, description, nil, action)
}

// TaskWithDeps registers a new task with dependencies.
func (e *Engine) TaskWithDeps(name, description string, deps []string, action func(ctx *Context) error) {
	e.Tasks[name] = &Task{
		Name:        name,
		Description: description,
		Action:      action,
		DependsOn:   deps,
	}
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
	ctx := &Context{
		Engine: e,
		Args:   os.Args[2:],
	}

	running := make(map[string]bool)
	if err := e.runTask(taskName, ctx, running); err != nil {
		fmt.Printf("Execution failed: %v\n", err)
		os.Exit(1)
	}
}

func (e *Engine) runTask(name string, ctx *Context, running map[string]bool) error {
	if e.executedTasks[name] {
		return nil
	}

	if running[name] {
		return fmt.Errorf("circular dependency detected: %s", name)
	}

	task, ok := e.Tasks[name]
	if !ok {
		return fmt.Errorf("unknown task: %s", name)
	}

	running[name] = true
	defer func() { running[name] = false }()

	// Run dependencies first
	for _, depName := range task.DependsOn {
		if err := e.runTask(depName, ctx, running); err != nil {
			return err
		}
	}

	// Run the task itself
	if err := task.Action(ctx); err != nil {
		return fmt.Errorf("task '%s' failed: %w", name, err)
	}

	e.executedTasks[name] = true
	return nil
}

func (e *Engine) PrintHelp() {
	fmt.Println("Usage: gobake <task> [args]")
	fmt.Println("\nAvailable tasks:")
	for name, task := range e.Tasks {
		fmt.Printf("  %-15s %s\n", name, task.Description)
	}
}
