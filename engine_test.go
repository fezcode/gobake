package gobake

import (
	"os"
	"testing"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	if e.Tasks == nil {
		t.Error("Tasks map should not be nil")
	}
}

func TestTaskRegistration(t *testing.T) {
	e := NewEngine()
	taskName := "test-task"
	taskDesc := "A test task"
	e.Task(taskName, taskDesc, func(ctx *Context) error {
		return nil
	})

	task, ok := e.Tasks[taskName]
	if !ok {
		t.Fatalf("Task %s not found", taskName)
	}
	if task.Name != taskName {
		t.Errorf("Expected name %s, got %s", taskName, task.Name)
	}
	if task.Description != taskDesc {
		t.Errorf("Expected description %s, got %s", taskDesc, task.Description)
	}
}

func TestLoadRecipeInfo(t *testing.T) {
	pimlContent := `(name) test-project
(version) 1.2.3
(description) Test description
(tools)
    > tool1
    > tool2`
	
	tmpFile := "test_recipe.piml"
	err := os.WriteFile(tmpFile, []byte(pimlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp piml file: %v", err)
	}
	defer os.Remove(tmpFile)

	e := NewEngine()
	err = e.LoadRecipeInfo(tmpFile)
	if err != nil {
		t.Fatalf("LoadRecipeInfo failed: %v", err)
	}

	if e.Info.Name != "test-project" {
		t.Errorf("Expected name test-project, got %s", e.Info.Name)
	}
	if e.Info.Version != "1.2.3" {
		t.Errorf("Expected version 1.2.3, got %s", e.Info.Version)
	}
	if len(e.Info.Tools) != 2 || e.Info.Tools[0] != "tool1" || e.Info.Tools[1] != "tool2" {
		t.Errorf("Tools not loaded correctly: %v", e.Info.Tools)
	}
}

func TestSaveRecipeInfo(t *testing.T) {
	e := NewEngine()
	e.Info = &RecipeInfo{
		Name:    "save-test",
		Version: "1.0.0",
		Tools:   []string{"github.com/a/b@latest"},
	}

	tmpFile := "test_save_recipe.piml"
	defer os.Remove(tmpFile)

	err := e.SaveRecipeInfo(tmpFile)
	if err != nil {
		t.Fatalf("SaveRecipeInfo failed: %v", err)
	}

	// Load it back to verify
	e2 := NewEngine()
	err = e2.LoadRecipeInfo(tmpFile)
	if err != nil {
		t.Fatalf("LoadRecipeInfo failed to read saved file: %v", err)
	}

	if e2.Info.Name != e.Info.Name {
		t.Errorf("Expected name %s, got %s", e.Info.Name, e2.Info.Name)
	}
	if e2.Info.Version != e.Info.Version {
		t.Errorf("Expected version %s, got %s", e.Info.Version, e2.Info.Version)
	}
	if len(e2.Info.Tools) != 1 || e2.Info.Tools[0] != e.Info.Tools[0] {
		t.Errorf("Expected tools %v, got %v", e.Info.Tools, e2.Info.Tools)
	}
}

func TestTaskDependencies(t *testing.T) {
	e := NewEngine()
	var order []string

	e.Task("a", "Task A", func(ctx *Context) error {
		order = append(order, "a")
		return nil
	})

	e.TaskWithDeps("b", "Task B", []string{"a"}, func(ctx *Context) error {
		order = append(order, "b")
		return nil
	})

	e.TaskWithDeps("c", "Task C", []string{"b", "a"}, func(ctx *Context) error {
		order = append(order, "c")
		return nil
	})

	// We can't use e.Execute() because it reads os.Args
	// We call runTask directly for testing
	ctx := &Context{Engine: e}
	running := make(map[string]bool)
	err := e.runTask("c", ctx, running)
	if err != nil {
		t.Fatalf("runTask failed: %v", err)
	}

	expected := []string{"a", "b", "c"}
	if len(order) != len(expected) {
		t.Fatalf("Expected order %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("At index %d: expected %s, got %s", i, v, order[i])
		}
	}
}

func TestCircularDependency(t *testing.T) {
	e := NewEngine()
	e.TaskWithDeps("a", "Task A", []string{"b"}, func(ctx *Context) error { return nil })
	e.TaskWithDeps("b", "Task B", []string{"a"}, func(ctx *Context) error { return nil })

	ctx := &Context{Engine: e}
	running := make(map[string]bool)
	err := e.runTask("a", ctx, running)
	if err == nil {
		t.Fatal("Expected error for circular dependency, got nil")
	}
	if err.Error() != "circular dependency detected: a" && err.Error() != "circular dependency detected: b" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := &Context{Engine: NewEngine()}
	tmpDir := "test_helpers_dir"
	defer os.RemoveAll(tmpDir)

	// Test Mkdir
	if err := ctx.Mkdir(tmpDir); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Fatal("Directory was not created")
	}

	// Test Copy
	srcFile := "test_src.txt"
	dstFile := "test_dst.txt"
	content := "hello world"
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create src file: %v", err)
	}
	defer os.Remove(srcFile)
	defer os.Remove(dstFile)

	if err := ctx.Copy(srcFile, dstFile); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
	gotContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read dst file: %v", err)
	}
	if string(gotContent) != content {
		t.Errorf("Expected content %s, got %s", content, string(gotContent))
	}

	// Test Remove
	if err := ctx.Remove(srcFile); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Fatal("File was not removed")
	}
}

func TestContextSetEnv(t *testing.T) {
	ctx := &Context{Engine: NewEngine()}
	key := "GOBAKE_TEST_VAR"
	val := "scoped-value"

	ctx.SetEnv(key, val)

	// Check that it's NOT in the global environment
	if os.Getenv(key) == val {
		t.Errorf("Env var %s should not be set globally", key)
	}

	// We can't easily test if it's passed to a sub-process without running one,
	// but we can verify it's in ctx.Env
	found := false
	expected := key + "=" + val
	for _, e := range ctx.Env {
		if e == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find %s in ctx.Env", expected)
	}
}
