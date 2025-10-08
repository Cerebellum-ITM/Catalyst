package core

import (
	"fmt"

	"catalyst/internal/app/styles"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type ClosePopupMsg struct{}

type PopupModel struct {
	Title      string
	Message    string
	Width      int
	Height     int
	ConfirmCmd tea.Cmd
	theme      *styles.Theme
}

func NewPopup(title, message string, confirmCmd tea.Cmd, theme *styles.Theme, width, height int) PopupModel {
	return PopupModel{
		Title:      title,
		Message:    message,
		Width:      width,
		Height:     height,
		ConfirmCmd: confirmCmd,
		theme:      theme,
	}
}

func (m *PopupModel) Init() tea.Cmd {
	return nil
}

func (m *PopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return ClosePopupMsg{} }
		case "enter":
			return m, tea.Batch(m.ConfirmCmd, func() tea.Msg { return ClosePopupMsg{} })
		}
	}
	return m, nil
}

func (m *PopupModel) View() string {
	content := fmt.Sprintf("%s\n\n%s", m.Message, "(enter to confirm, esc to cancel)")

	popupBox := lipgloss.NewStyle().
		Width(m.Width/2).
		Align(lipgloss.Center).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Render(content)

	return popupBox
}
