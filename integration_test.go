package gobake_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Version(t *testing.T) {
	// Build the binary
	binPath := filepath.Join(os.TempDir(), "gobake_test_bin.exe")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/gobake")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build gobake: %v", err)
	}
	defer os.Remove(binPath)

	cmd := exec.Command(binPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gobake version failed: %v", err)
	}

	if !strings.Contains(string(output), "gobake version") {
		t.Errorf("Expected version output, got: %s", string(output))
	}
}

func TestCLI_Bump(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "gobake_bump_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy recipe.piml
	pimlPath := filepath.Join(tmpDir, "recipe.piml")
	initialContent := `(version) 1.0.0` + "\n"
	if err := os.WriteFile(pimlPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create recipe.piml: %v", err)
	}

	// Build gobake
	binPath := filepath.Join(os.TempDir(), "gobake_test_bin_bump.exe")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/gobake")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build gobake: %v", err)
	}
	defer os.Remove(binPath)

	// Run gobake bump patch
	cmd := exec.Command(binPath, "bump", "patch")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gobake bump patch failed: %v\nOutput: %s", err, string(output))
	}

	// Check result
	updatedContent, err := os.ReadFile(pimlPath)
	if err != nil {
		t.Fatalf("Failed to read updated recipe.piml: %v", err)
	}

	if !strings.Contains(string(updatedContent), "(version) 1.0.1") {
		t.Errorf("Expected version 1.0.1, got: %s", string(updatedContent))
	}
}

func TestCLI_Init(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "gobake_init_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build gobake
	binPath := filepath.Join(os.TempDir(), "gobake_test_bin_init.exe")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/gobake")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build gobake: %v", err)
	}
	defer os.Remove(binPath)

	// Run gobake init
	cmd := exec.Command(binPath, "init")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("gobake init failed: %v\nOutput: %s", err, string(output))
	}

	// 1. Check recipe.piml
	pimlPath := filepath.Join(tmpDir, "recipe.piml")
	if _, err := os.Stat(pimlPath); os.IsNotExist(err) {
		t.Errorf("recipe.piml was not created")
	}

	// 2. Check Recipe.go
	recipePath := filepath.Join(tmpDir, "Recipe.go")
	if _, err := os.Stat(recipePath); os.IsNotExist(err) {
		t.Errorf("Recipe.go was not created")
	}

	// 3. Check go.mod
	modPath := filepath.Join(tmpDir, "go.mod")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		t.Errorf("go.mod was not created")
	}

	// 4. Verify gobake dependency in go.mod
	modContent, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(modContent), "github.com/fezcode/gobake") {
		t.Errorf("Expected gobake dependency in go.mod, but it was missing.\nContent: %s", string(modContent))
	}
}

