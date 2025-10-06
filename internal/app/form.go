package app

import (
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// updateRuneForm contains the shared logic for both creating and editing rune forms.
func updateRuneForm(msg tea.Msg, m *Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return tea.Batch(cmds...)

		case "enter":
			// In command fields (index 2+), enter creates a new field
			if m.focusIndex >= 2 {
				newInput := textinput.New()
				newInput.Placeholder = "Command"
				newInput.Focus()
				
				newIndex := m.focusIndex + 1
				// Insert new input after the current one
				m.inputs = append(m.inputs[:newIndex], append([]textinput.Model{newInput}, m.inputs[newIndex:]...)...)
				
				m.inputs[m.focusIndex].Blur()
				m.focusIndex = newIndex
				return textinput.Blink
			}

		case "backspace":
			// In command fields (index > 2), backspace on empty field deletes it
			if m.focusIndex > 2 && m.inputs[m.focusIndex].Value() == "" {
				m.inputs = append(m.inputs[:m.focusIndex], m.inputs[m.focusIndex+1:]...)
				m.focusIndex--
				m.inputs[m.focusIndex].Focus()
				return textinput.Blink
			}
		}
	}

	// Update the focused text input
	cmd := m.updateInputs(msg)
	return cmd
}
