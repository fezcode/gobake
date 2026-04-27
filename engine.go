package gobake

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"sort"
	"strings"
)

var Version = "0.3.0"

func init() {
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				Version = info.Main.Version
			}
		}
	}
}

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
	Env    []string
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
	reserved := map[string]bool{
		"init":        true,
		"version":     true,
		"bump":        true,
		"template":    true,
		"add-tool":    true,
		"remove-tool": true,
		"add-dep":     true,
		"remove-dep":  true,
		"help":        true,
	}
	if reserved[name] {
		fmt.Printf("Error: Task name '%s' is reserved by gobake CLI.\n", name)
		os.Exit(1)
	}

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
	cmd.Env = append(cmd.Env, ctx.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Log prints a formatted message to stdout.
func (ctx *Context) Log(format string, a ...interface{}) {
	fmt.Printf("[gobake] "+format+"\n", a...)
}

// Mkdir creates a directory and any necessary parents.
func (ctx *Context) Mkdir(path string) error {
	ctx.Log("Creating directory: %s", path)
	return os.MkdirAll(path, 0755)
}

// Remove removes a file or directory.
func (ctx *Context) Remove(path string) error {
	ctx.Log("Removing: %s", path)
	return os.RemoveAll(path)
}

// Copy copies a file from src to dst.
func (ctx *Context) Copy(src, dst string) error {
	ctx.Log("Copying %s -> %s", src, dst)
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// SetEnv sets an environment variable for the current task context.
func (ctx *Context) SetEnv(key, value string) {
	ctx.Env = append(ctx.Env, fmt.Sprintf("%s=%s", key, value))
}

// Run executes a shell command and waits for it to finish.
func (ctx *Context) Run(name string, args ...string) error {
	return ctx.RunIn("", name, args...)
}

// RunIn executes a shell command in the given working directory.
// An empty dir uses the current working directory.
func (ctx *Context) RunIn(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), ctx.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunOutput executes a command and returns its captured stdout.
// Stderr is still streamed to os.Stderr so failures stay visible.
func (ctx *Context) RunOutput(name string, args ...string) (string, error) {
	return ctx.RunInOutput("", name, args...)
}

// RunInOutput is RunOutput with an explicit working directory.
func (ctx *Context) RunInOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), ctx.Env...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\r\n"), err
}

// Execute runs the engine based on command line arguments.
//
// Leading arguments that name registered tasks are all executed in order.
// The first argument that is not a task name (and everything after it) is
// passed through to the last task as ctx.Args. So:
//
//	gobake build           -> runs build
//	gobake build test      -> runs build then test
//	gobake build foo.txt   -> runs build with Args=["foo.txt"]
//	gobake build test x y  -> runs build then test with Args=["x", "y"]
func (e *Engine) Execute() {
	if len(os.Args) < 2 {
		e.PrintHelp()
		return
	}

	args := os.Args[1:]
	var taskNames []string
	var trailingArgs []string
	for i, a := range args {
		if _, ok := e.Tasks[a]; ok {
			taskNames = append(taskNames, a)
			continue
		}
		trailingArgs = args[i:]
		break
	}

	if len(taskNames) == 0 {
		fmt.Printf("Unknown task: %s\n", args[0])
		e.PrintHelp()
		os.Exit(1)
	}

	ctx := &Context{
		Engine: e,
		Args:   trailingArgs,
	}

	running := make(map[string]bool)
	for _, name := range taskNames {
		if err := e.runTask(name, ctx, running); err != nil {
			fmt.Printf("Execution failed: %v\n", err)
			os.Exit(1)
		}
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
	fmt.Println("Usage: gobake <task> [<task>...] [args]")
	fmt.Println("\nAvailable tasks:")
	names := make([]string, 0, len(e.Tasks))
	for name := range e.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("  %-15s %s\n", name, e.Tasks[name].Description)
	}
}
