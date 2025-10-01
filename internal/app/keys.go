package app

import "github.com/charmbracelet/bubbles/v2/key"

// KeyMap defines a set of keybindings.
// It implements the help.KeyMap interface.
type KeyMap struct {
	Enter      key.Binding
	Delete     key.Binding
	Quit       key.Binding
	GlobalQuit key.Binding
	Help       key.Binding
	Esc        key.Binding
}

func mainListKeys() KeyMap {
	return KeyMap{
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "Delete")),
		Quit:       key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		GlobalQuit: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	b := []key.Binding{}
	if k.Enter.Enabled() {
		b = append(b, k.Enter)
	}
	if k.Quit.Enabled() {
		b = append(b, k.Quit)
	}
	if k.Delete.Enabled() {
		b = append(b, k.Delete)
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
	if k.Quit.Enabled() {
		b = append(b, k.Quit)
	}
	if k.Delete.Enabled() {
		b = append(b, k.Delete)
	}
	if k.GlobalQuit.Enabled() {
		b = append(b, k.GlobalQuit)
	}
	if k.Help.Enabled() {
		b = append(b, k.Help)
	}
	return [][]key.Binding{b}
}
