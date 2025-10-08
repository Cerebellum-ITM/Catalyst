package styles

import (
	"image/color"
	"time"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/bubbles/v2/textinput"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type Theme struct {
	Name   string
	IsDark bool
	Logo   color.Color

	FgBase      color.Color
	FgMuted     color.Color
	FgHalfMuted color.Color
	FgSubtle    color.Color

	BorderFocus      color.Color
	FillTextLine     color.Color
	FocusableElement color.Color

	BgOverlay color.Color
	Input     color.Color
	Output    color.Color

	Primary   color.Color
	Secondary color.Color
	Tertiary  color.Color
	Accent    color.Color
	Blur      color.Color

	Success color.Color
	Error   color.Color
	Warning color.Color
	Info    color.Color
	Fatal   color.Color

	Yellow color.Color
	Purple color.Color
	White  color.Color
	Red    color.Color
	Green  color.Color
	Black  color.Color

	styles *Styles
}

type Styles struct {
	Base        lipgloss.Style
	HeaderStyle lipgloss.Style
	FooterStyle lipgloss.Style
	LineStyle   lipgloss.Style
	Help        help.Styles
	TextArea    textarea.Styles
	Textinput   textinput.Styles
}

func (t *Theme) AppStyles() *Styles {
	if t.styles == nil {
		t.styles = t.buildStyles()
	}
	return t.styles
}

func (t *Theme) buildStyles() *Styles {
	base := lipgloss.NewStyle().
		Foreground(t.FgBase)
	HeaderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(false).
		Padding(0, 1)
	FooterStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(false).
		Padding(0, 1)

	return &Styles{
		Base:        base,
		HeaderStyle: HeaderStyle,
		FooterStyle: FooterStyle,
		LineStyle:   lipgloss.NewStyle(),
		Textinput: textinput.Styles{
			Focused: textinput.StyleState{
				Text:        base.Foreground(t.Accent),
				Placeholder: base.Foreground(t.Blur),
				Prompt:      base.Foreground(t.Yellow),
			},
			Blurred: textinput.StyleState{
				Text:        base.Foreground(t.Black),
				Placeholder: base.Foreground(t.Black),
				Prompt:      base.Foreground(t.Black),
			},
			Cursor: textinput.CursorStyle{
				Color:      t.Accent,
				Shape:      tea.CursorBar,
				Blink:      true,
				BlinkSpeed: time.Millisecond * 500,
			},
		},
		Help: help.Styles{
			ShortKey:       base.Foreground(t.Accent),
			ShortDesc:      base.Foreground(t.FgMuted),
			ShortSeparator: base.Foreground(t.White),
			FullKey:        base.Foreground(t.Accent),
			FullDesc:       base.Foreground(t.FgMuted),
			FullSeparator:  base.Foreground(t.White),
			Ellipsis:       base.Foreground(t.FgSubtle),
		},
	}
}
