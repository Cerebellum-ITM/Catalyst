package core

import (
	"catalyst/internal/app/styles"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// CustomTextInput wraps textinput.Model to provide custom styling.
type CustomTextInput struct {
	textinput.Model
	Theme styles.Theme
}

// NewTextInput creates a new CustomTextInput.
func NewTextInput(theme styles.Theme) CustomTextInput {
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.VirtualCursor = true
	ti.Styles = theme.AppStyles().Textinput
	return CustomTextInput{
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
		cti.Prompt = "❯ "
	} else {
		cti.Prompt = "  "
	}
	return cti.Model.View()
}
