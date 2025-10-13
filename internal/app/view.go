package app

import (
	"fmt"
	// "os"
	"image/color"
	"strings"

	"catalyst/internal/ascii"
	"catalyst/internal/types"
	"catalyst/internal/utils"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	focusColor      color.Color
	focusColorText  color.Color
	blurColor       color.Color
	HeaderStyle     lipgloss.Style
	FooterStyle     lipgloss.Style
	LineStyle       lipgloss.Style
	highlight       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	VerticalSpace   = lipgloss.NewStyle().Height(1).Render("")
	HorizontalSpace = lipgloss.NewStyle().Width(10).Render("")
)

type BorderAlignment int

const (
	AlignHeader BorderAlignment = iota
	AlignFooter
)

func (m *Model) setAppStyles() {
	HeaderStyle = m.Theme.AppStyles().HeaderStyle
	FooterStyle = m.Theme.AppStyles().FooterStyle
	LineStyle = m.Theme.AppStyles().LineStyle
}

func (m *Model) setColorVariables(state string) (textColor, lineColor ansi.Color) {
	focusColor = m.Theme.Primary
	focusColorText = m.Theme.Accent
	blurColor = m.Theme.Blur
	if state == "focus" {
		textColor = focusColorText
		lineColor = focusColor
	} else {
		textColor = blurColor
		lineColor = blurColor
	}
	return textColor, lineColor
}

func (m *Model) buildStyledBorder(
	state string,
	content string,
	baseStyle lipgloss.Style,
	componentWidth int,
	alignment BorderAlignment,
) string {
	textColor, lineColor := m.setColorVariables(state)

	styledContent := baseStyle.Foreground(textColor).Render(content)

	contentWidth := lipgloss.Width(styledContent)
	line := LineStyle.Foreground(lineColor).
		Render(strings.Repeat("─", max(0, componentWidth-contentWidth)))

	switch alignment {
	case AlignHeader:
		return lipgloss.JoinHorizontal(lipgloss.Left, styledContent, line)
	case AlignFooter:
		return lipgloss.JoinHorizontal(lipgloss.Left, line, styledContent)
	default:
		return ""
	}
}

func (m *Model) mainMenuHeaderViewport(state string) string {
	title := fmt.Sprintf("Spellbook: %s", utils.TruncatePath(m.pwd, 1))
	return m.buildStyledBorder(
		state,
		title,
		HeaderStyle,
		m.viewportSpellBook.Width(),
		AlignHeader,
	)
}

func (m *Model) mainMenuFooterVierPort(state string) string {
	info := fmt.Sprintf("%3.f%%", m.viewportSpellBook.ScrollPercent()*100)
	return m.buildStyledBorder(
		state,
		info,
		FooterStyle,
		m.viewportSpellBook.Width(),
		AlignFooter,
	)
}

func (m *Model) mainMenuHeaderList(state string) string {
	return m.buildStyledBorder(
		state,
		"Select an option",
		HeaderStyle,
		m.menuItems.Width(),
		AlignHeader,
	)
}

func (m *Model) mainMenuFooterList(state string) string {
	return m.buildStyledBorder(
		state,
		"",
		FooterStyle,
		m.menuItems.Width(),
		AlignFooter,
	)
}

func (m *Model) runesListHeader(state string) string {
	return m.buildStyledBorder(
		state,
		"Runes",
		HeaderStyle,
		m.runesList.Width(),
		AlignHeader,
	)
}

func (m *Model) runesListFooter(state string) string {
	return m.buildStyledBorder(
		state,
		"",
		FooterStyle,
		m.runesList.Width(),
		AlignFooter,
	)
}

func (m *Model) runeDetailHeader(state string) string {
	title := "Rune Details"
	return m.buildStyledBorder(
		state,
		title,
		HeaderStyle,
		m.viewportSpellBook.Width(),
		AlignHeader,
	)
}

func (m *Model) runeDetailFooter(state string) string {
	info := fmt.Sprintf("%3.f%%", m.viewportSpellBook.ScrollPercent()*100)
	return m.buildStyledBorder(
		state,
		info,
		FooterStyle,
		m.viewportSpellBook.Width(),
		AlignFooter,
	)
}

func (m *Model) formHeader(state string) string {
	title := "Create New Rune"
	// In a real implementation, you might check if you are editing vs creating
	// and change the title accordingly.
	return m.buildStyledBorder(
		state,
		title,
		HeaderStyle,
		(m.width / 2), // Assuming form takes half the width
		AlignHeader,
	)
}

func (m *Model) formFooter(state string) string {
	return m.buildStyledBorder(
		state,
		"",
		FooterStyle,
		(m.width / 2),
		AlignFooter,
	)
}

func (m *Model) previewHeader(state string) string {
	return m.buildStyledBorder(
		state,
		"Live Preview",
		HeaderStyle,
		(m.width / 2),
		AlignHeader,
	)
}

func (m *Model) previewFooter(state string) string {
	info := fmt.Sprintf("%3.f%%", m.formViewport.ScrollPercent()*100)
	return m.buildStyledBorder(
		state,
		info,
		FooterStyle,
		(m.width / 2),
		AlignFooter,
	)
}

func (m *Model) executingRuneHeaderLeft(state string) string {
	return m.buildStyledBorder(
		state,
		"Commands",
		HeaderStyle,
		(m.width/3),
		AlignHeader,
	)
}

func (m *Model) executingRuneFooterLeft(state string) string {
	return m.buildStyledBorder(
		state,
		"",
		FooterStyle,
		(m.width/3),
		AlignFooter,
	)
}

func (m *Model) executingRuneHeaderRight(state string) string {
	return m.buildStyledBorder(
		state,
		"Output",
		HeaderStyle,
		(m.width*2/3),
		AlignHeader,
	)
}

func (m *Model) executingRuneFooterRight(state string) string {
	info := fmt.Sprintf("%3.f%%", m.executingViewport.ScrollPercent()*100)
	return m.buildStyledBorder(
		state,
		info,
		FooterStyle,
		(m.width*2/3),
		AlignFooter,
	)
}

func formatRuneDetail(rune types.Rune) string {
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# %s\n", rune.Name))
	md.WriteString(fmt.Sprintf("# %s\n", "Description"))
	md.WriteString(fmt.Sprintf("> %s\n\n", rune.Description))
	md.WriteString("```sh\n")
	for _, cmd := range rune.Commands {
		md.WriteString(fmt.Sprintf("%s\n", cmd))
	}
	md.WriteString("```\n")
	return md.String()
}

func (m Model) showProntMessage(availableHeightForMainContent int) string {
	var prontMessage string
	asciiLogo := ascii.PrintLogo()
	specsText := ascii.PrintSpecs()

	finalPrompt := lipgloss.JoinHorizontal(
		lipgloss.Left,
		HorizontalSpace,
		asciiLogo,
		HorizontalSpace,
		specsText,
	)
	prontMessage = lipgloss.Place(
		m.width,
		availableHeightForMainContent,
		lipgloss.Left,
		lipgloss.Center,
		finalPrompt,
	)

	return prontMessage
}

func (m *Model) View() string {
	var s strings.Builder
	var stateView string

	statusBarContent := m.StatusBar.Render()
	helpView := lipgloss.NewStyle().Padding(0, 2).SetString(m.help.View(m.keys)).String()
	printMessage := m.showProntMessage(m.availableHeight)
	m.setAppStyles()

	switch m.state {
	case checkingSpellbook:
		m.StatusBar.Content = "Checking for Spellbook..."
		s.WriteString(printMessage)
	case creatingSpellbook:
		m.StatusBar.Content = "Creating Spellbook..."
		s.WriteString(printMessage)
	case spellbookLoaded:
		s.WriteString(printMessage)
	case ready:
		var (
			listElementState     = "blur"
			viewportElementState = "blur"
		)

		glamourStyle := styles.DarkStyleConfig
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithStyles(glamourStyle),
			glamour.WithWordWrap(m.viewportSpellBook.Width()),
		)
		glamourContent, _ := renderSpellbook(m.spellbook)
		glamourContentStr, _ := renderer.Render(glamourContent)

		m.viewportSpellBook.SetContent(glamourContentStr)

		switch m.focusedElement {
		case listElement:
			listElementState = "focus"
		case viewportElement:
			viewportElementState = "focus"
		}

		menuHeaderVpContent := m.mainMenuHeaderViewport(viewportElementState)
		menuFooterVpContent := m.mainMenuFooterVierPort(viewportElementState)
		extraContent := lipgloss.Height(menuHeaderVpContent + menuFooterVpContent)
		m.recalculateSizes(WithExtraContent(extraContent))

		rightSideContent := lipgloss.JoinVertical(
			lipgloss.Left,
			menuHeaderVpContent,
			m.viewportSpellBook.View(),
			menuFooterVpContent,
		)

		leftSideContent := lipgloss.JoinVertical(
			lipgloss.Left,
			m.mainMenuHeaderList(listElementState),
			m.menuItems.View(),
			m.mainMenuFooterList(listElementState),
		)

		stateView = lipgloss.JoinHorizontal(
			lipgloss.Left,
			leftSideContent,
			rightSideContent,
		)
		s.WriteString(stateView)

	case showingRunes:
		var (
			listElementState     = "blur"
			viewportElementState = "blur"
		)

		switch m.focusedElement {
		case listElement:
			listElementState = "focus"
		case viewportElement:
			viewportElementState = "focus"
		}

		// Since this view has headers and footers, we need to account for their height
		runeListHeader := m.runesListHeader(listElementState)
		runeListFooter := m.runesListFooter(listElementState)
		runeDetailHeader := m.runeDetailHeader(viewportElementState)
		runeDetailFooter := m.runeDetailFooter(viewportElementState)

		// Calculate height for components, subtracting the borders
		extraContentHeight := lipgloss.Height(runeListHeader) + lipgloss.Height(runeListFooter)
		m.recalculateSizes(WithExtraContent(extraContentHeight))

		leftSideContent := lipgloss.JoinVertical(
			lipgloss.Left,
			runeListHeader,
			m.runesList.View(),
			runeListFooter,
		)

		rightSideContent := lipgloss.JoinVertical(
			lipgloss.Left,
			runeDetailHeader,
			m.viewportSpellBook.View(),
			runeDetailFooter,
		)

		stateView = lipgloss.JoinHorizontal(
			lipgloss.Left,
			leftSideContent,
			rightSideContent,
		)
		s.WriteString(stateView)

	case creatingRune, editingRune:
		var (
			formState    = "blur"
			previewState = "blur"
			formBuilder  strings.Builder
		)

		switch m.focusedElement {
		case formElement:
			formState = "focus"
		case viewportElement:
			previewState = "focus"
		}

		// Build the form side content
		for i := range m.inputs {
			formBuilder.WriteString(m.inputs[i].View() + "\n")
		}
		submitButton := "  Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		formBuilder.WriteString(fmt.Sprintf("\n%s\n", submitButton))

		formHeader := m.formHeader(formState)
		formFooter := m.formFooter(formState)
		previewHeader := m.previewHeader(previewState)
		previewFooter := m.previewFooter(previewState)

		// Correctly calculate available height *before* building the final layout
		extraContentHeight := lipgloss.Height(formHeader) + lipgloss.Height(formFooter)
		m.recalculateSizes(WithExtraContent(extraContentHeight))

		// Now calculate the spacer with the correct available height
		formContent := formBuilder.String()
		formHeight := lipgloss.Height(formContent)
		spacerHeight := max(0, m.availableHeight-formHeight)
		spacer := lipgloss.NewStyle().Height(spacerHeight).Render("")

		leftSideContent := lipgloss.JoinVertical(
			lipgloss.Left,
			formHeader,
			formContent,
			spacer,
			formFooter,
		)

		rightSideContent := lipgloss.JoinVertical(
			lipgloss.Left,
			previewHeader,
			m.formViewport.View(),
			previewFooter,
		)

		stateView = lipgloss.JoinHorizontal(
			lipgloss.Left,
			leftSideContent,
			rightSideContent,
		)
		s.WriteString(stateView)

	case executingRune:
		headerLeft := m.executingRuneHeaderLeft("blur")
		footerLeft := m.executingRuneFooterLeft("blur")
		headerRight := m.executingRuneHeaderRight("focus")
		footerRight := m.executingRuneFooterRight("focus")
		extraContentHeight := lipgloss.Height(headerRight) + lipgloss.Height(footerRight)
		m.recalculateSizes(WithExtraContent(extraContentHeight))

		// For now, the left side is empty, but we set it up for future use.
		var leftSideContent string
		if m.logsView != nil {
			leftSideContent = lipgloss.JoinVertical(
				lipgloss.Left,
				headerLeft,
				m.logsView.View(),
				footerLeft,
			)
		} else {
			leftSideContent = lipgloss.JoinVertical(
				lipgloss.Left,
				headerLeft,
				lipgloss.NewStyle().Height(m.availableHeight).Render(""), // Empty for now
				footerLeft,
			)
		}

		m.executingViewport.SetContent(m.output)

		rightSideContent := lipgloss.JoinVertical(
			lipgloss.Left,
			headerRight,
			m.executingViewport.View(),
			footerRight,
		)

		stateView = lipgloss.JoinHorizontal(
			lipgloss.Left,
			leftSideContent,
			rightSideContent,
		)
		s.WriteString(stateView)

	case showingLoegs:
		s.WriteString("Loegs (Environment Variables):\n\n")
		if len(m.loegKeys) == 0 {
			s.WriteString("No loegs found.\n")
		} else {
			for i, k := range m.loegKeys {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s = %s\n", highlight.Render(cursor), k, m.spellbook.Loegs[k]))
			}
		}

	case creatingLoeg:
		s.WriteString("Create a New Loeg\n\n")
		for i := range m.inputs {
			s.WriteString(m.inputs[i].View() + "\n")
		}

		submitButton := "Submit"
		if m.focusIndex == len(m.inputs) {
			submitButton = highlight.Render(submitButton)
		}
		s.WriteString(fmt.Sprintf("\n%s\n", submitButton))

	case errState:
		s.WriteString(fmt.Sprintf("An error occurred: %v\n\n", m.err))

	case showingHistory:
		s.WriteString("Execution History:\n\n")
		if len(m.history) == 0 {
			s.WriteString("No history found.\n")
		} else {
			for i, entry := range m.history {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s on %s at %s\n",
					highlight.Render(cursor),
					entry.RuneID,
					entry.SpellbookID,
					entry.ExecutedAt.Format("2006-01-02 15:04:05")),
				)
			}
		}

	}

	uiElements := s.String()

	if m.lockScreen != nil {
		uiElements = m.lockScreen.View()
	}

	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		statusBarContent,
		VerticalSpace,
		uiElements,
		helpView,
	)

	mainLayer := lipgloss.NewLayer(mainContent)
	canvas := lipgloss.NewCanvas(mainLayer)

	if m.popup != nil {
		popupView := m.popup.View()
		popupWidth := lipgloss.Width(popupView)
		popupHeight := lipgloss.Height(popupView)
		startX := (m.width - popupWidth) / 2
		startY := (m.height - popupHeight) / 2
		popupLayer := lipgloss.NewLayer(popupView).X(startX).Y(startY)
		canvas = lipgloss.NewCanvas(mainLayer, popupLayer)
	}

	return canvas.Render()
}

func renderSpellbook(sb *Spellbook) (string, error) {
	if sb == nil {
		return "", fmt.Errorf("spellbook is nil")
	}

	var md strings.Builder

	// Title
	// md.WriteString(fmt.Sprintf("# %s\n\n", sb.Name))

	// Runes Section
	md.WriteString("# Runes\n\n")
	if len(sb.Runes) == 0 {
		md.WriteString("No runes found.\n\n")
	} else {
		for _, r := range sb.Runes {
			md.WriteString(fmt.Sprintf("## %s\n", r.Name))
			md.WriteString(fmt.Sprintf("> %s\n\n", r.Description))
			md.WriteString("```sh\n")
			for _, cmd := range r.Commands {
				md.WriteString(fmt.Sprintf("%s\n", cmd))
			}
			md.WriteString("```\n\n")
		}
	}

	// Loegs Section
	// md.WriteString("## Loegs (Environment Variables)\n\n")
	// if len(sb.Loegs) == 0 {
	// 	md.WriteString("No loegs found.\n\n")
	// } else {
	// 	md.WriteString("| Key | Value |\n")
	// 	md.WriteString("|-----|-------|\n")
	// 	for k, v := range sb.Loegs {
	// 		md.WriteString(fmt.Sprintf("| %s | %s |\n", k, v))
	// 	}
	// }

	// Write to file for debugging
	// err := os.WriteFile("spellbook_output.md", []byte(md.String()), 0o644)
	// if err != nil {
	// return "", fmt.Errorf("failed to write spellbook to file: %w", err)
	// }

	return md.String(), nil
}
