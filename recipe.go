package gobake

import (
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
