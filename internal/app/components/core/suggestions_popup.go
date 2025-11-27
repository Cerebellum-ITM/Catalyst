package core

import (
	"fmt"

	"catalyst/internal/app/styles"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var VerticalSpace = lipgloss.NewStyle().Height(1).Render("")

type (
	CloseSuggestionsFinderPopupMsg struct {
		SuggestionStr  string
		CursorPosition int
	}
	// CustomTextInput wraps textinput.Model to provide custom styling.
	PopupSuggestionsFinder struct {
		textinput.Model
		Name          string
		Theme         styles.Theme
		Help          help.Model
		Width         int
		Height        int
		CursorPostion int
		keys          KeyMap
	}
)

// NewTextInput creates a new CustomTextInput.
func NewPopupSuggestionsFinder(
	Name string,
	theme styles.Theme,
	width, height, runeIndex int,
) *PopupSuggestionsFinder {
	ti := textinput.New()
	ti.Placeholder = "Start typing to search for commands or suggestions...."
	ti.ShowSuggestions = true
	ti.SetStyles(theme.AppStyles().Textinput)
	ti.Focus()
	ti.SetWidth(width / 2)

	help := help.New()
	help.Styles = theme.AppStyles().Help

	return &PopupSuggestionsFinder{
		Name:          Name,
		Model:         ti,
		Help:          help,
		Theme:         theme,
		Width:         width,
		Height:        height,
		CursorPostion: runeIndex,
		keys:          popupSuggestionsFinderKeys(),
	}
}

func (psf *PopupSuggestionsFinder) Init() tea.Cmd {
	return nil
}

func (psf *PopupSuggestionsFinder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		psf.Width = msg.Width
		psf.Height = msg.Height
		psf.Model.SetWidth(msg.Width / 2)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, psf.keys.Help):
			psf.Help.ShowAll = !psf.Help.ShowAll
			return psf, nil

		case key.Matches(msg, psf.keys.Quit):
			return psf, tea.Quit
		case key.Matches(msg, psf.keys.Esc):
			return psf, func() tea.Msg {
				return CloseSuggestionsFinderPopupMsg{SuggestionStr: "", CursorPosition: psf.CursorPostion}
			}
		case key.Matches(msg, psf.keys.Enter):
			return psf, func() tea.Msg {
				return CloseSuggestionsFinderPopupMsg{SuggestionStr: psf.Model.Value(), CursorPosition: psf.CursorPostion}
			}
		}
	}

	psf.Model, cmd = psf.Model.Update(msg)
	return psf, cmd
}

// View applies custom logic before rendering.
func (psf *PopupSuggestionsFinder) View() string {
	if psf.Focused() {
		psf.Prompt = fmt.Sprintf("❯ %s: ", psf.Name)
	} else {
		psf.Prompt = fmt.Sprintf("  %s: ", psf.Name)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		psf.Model.View(),
		VerticalSpace,
		lipgloss.NewStyle().Padding(0, 2).SetString(psf.Help.View(psf.keys)).String(),
	)

	popupBox := lipgloss.NewStyle().
		Width(psf.Width/2).
		Align(lipgloss.Center).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(psf.Theme.Accent).
		Render(content)

	return popupBox
}
