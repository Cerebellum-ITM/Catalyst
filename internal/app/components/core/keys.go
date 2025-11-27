package core

import "github.com/charmbracelet/bubbles/v2/key"

type KeyMap struct {
	// General
	Enter      key.Binding
	Quit       key.Binding
	GlobalQuit key.Binding
	Help       key.Binding
	Esc        key.Binding

	// popupSuggestionsFinder
	AcceptSuggestion key.Binding

	// Navigation
	NextSuggestion key.Binding
	PrevSuggestion key.Binding
}

func popupSuggestionsFinderKeys() KeyMap {
	return KeyMap{
		AcceptSuggestion: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "Accept suggestion"),
		),
		Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		NextSuggestion: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "Next suggestion"),
		),
		PrevSuggestion: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "Previous suggestion"),
		),
		Quit:       key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		GlobalQuit: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Esc:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	b := []key.Binding{}

	if k.Enter.Enabled() {
		b = append(b, k.Enter)
	}

	if k.AcceptSuggestion.Enabled() {
		b = append(b, k.AcceptSuggestion)
	}

	if k.Esc.Enabled() {
		b = append(b, k.Esc)
	}

	if k.NextSuggestion.Enabled() {
		b = append(b, k.NextSuggestion)
	}

	if k.GlobalQuit.Enabled() {
		b = append(b, k.GlobalQuit)
	}

	if k.Help.Enabled() {
		b = append(b, k.Help)
	}

	return b
}

func (k KeyMap) FullHelp() [][]key.Binding {
	b := []key.Binding{}
	if k.Enter.Enabled() {
		b = append(b, k.Enter)
	}

	if k.AcceptSuggestion.Enabled() {
		b = append(b, k.AcceptSuggestion)
	}

	if k.Esc.Enabled() {
		b = append(b, k.Esc)
	}

	if k.NextSuggestion.Enabled() {
		b = append(b, k.NextSuggestion)
	}
	if k.PrevSuggestion.Enabled() {
		b = append(b, k.PrevSuggestion)
	}

	if k.GlobalQuit.Enabled() {
		b = append(b, k.GlobalQuit)
	}

	if k.Help.Enabled() {
		b = append(b, k.Help)
	}

	return [][]key.Binding{b}
}
