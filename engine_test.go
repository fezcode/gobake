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
