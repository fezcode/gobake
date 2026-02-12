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
