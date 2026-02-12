package gobake

import (
	"fmt"
	"os"

	"github.com/fezcode/go-piml"
)

// LoadRecipeInfo reads and parses recipe.piml.
func (e *Engine) LoadRecipeInfo(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var info RecipeInfo
	if err := piml.Unmarshal(data, &info); err != nil {
		return err
	}

	e.Info = &info
	return nil
}

// SaveRecipeInfo saves the current RecipeInfo to a file in PIML format.
func (e *Engine) SaveRecipeInfo(path string) error {
	if e.Info == nil {
		return fmt.Errorf("no recipe info to save")
	}

	data, err := piml.Marshal(e.Info)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
