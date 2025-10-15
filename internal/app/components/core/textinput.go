package core

import (
	"fmt"

	"catalyst/internal/app/styles"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// CustomTextInput wraps textinput.Model to provide custom styling.
type CustomTextInput struct {
	textinput.Model
	Name  string
	Theme styles.Theme
}

// NewTextInput creates a new CustomTextInput.
func NewTextInput(Name string, theme styles.Theme) CustomTextInput {
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.ShowSuggestions = true
	ti.SetStyles(theme.AppStyles().Textinput)
	return CustomTextInput{
		Name:  Name,
		Model: ti,
		Theme: theme,
	}
}

func (cti CustomTextInput) Update(msg tea.Msg) (CustomTextInput, tea.Cmd) {
	var cmd tea.Cmd
	cti.Model, cmd = cti.Model.Update(msg)
	return cti, cmd
}

// View applies custom logic before rendering.
func (cti CustomTextInput) View() string {
	if cti.Focused() {
		cti.Prompt = fmt.Sprintf("❯ %s: ", cti.Name)
	} else {
		cti.Prompt = fmt.Sprintf("  %s: ", cti.Name)
	}
	return cti.Model.View()
}
