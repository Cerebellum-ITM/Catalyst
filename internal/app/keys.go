package app

import "github.com/charmbracelet/bubbles/v2/key"

// KeyMap defines a set of keybindings.
// It implements the help.KeyMap interface.
type KeyMap struct {
	// General
	Enter      key.Binding
	Quit       key.Binding
	GlobalQuit key.Binding
	Help       key.Binding
	Esc        key.Binding

	// Navigation
	Up   key.Binding
	Down key.Binding

	// Rune specific
	Edit   key.Binding
	Delete key.Binding
	New    key.Binding
}

func mainListKeys() KeyMap {
	return KeyMap{
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Quit:       key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		GlobalQuit: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func viewingRunesKeys() KeyMap {
	return KeyMap{
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "run")),
		Edit:       key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Esc:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		GlobalQuit: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func viewingLoegsKeys() KeyMap {
	return KeyMap{
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		New:        key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Esc:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		GlobalQuit: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func formKeys() KeyMap {
	return KeyMap{
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
		Esc:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		GlobalQuit: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func executingRuneKeys() KeyMap {
	return KeyMap{
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "back")),
		Esc:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		GlobalQuit: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	b := []key.Binding{}
	if k.Up.Enabled() {
		b = append(b, k.Up)
	}
	if k.Down.Enabled() {
		b = append(b, k.Down)
	}
	if k.Enter.Enabled() {
		b = append(b, k.Enter)
	}
	if k.New.Enabled() {
		b = append(b, k.New)
	}
	if k.Edit.Enabled() {
		b = append(b, k.Edit)
	}
	if k.Delete.Enabled() {
		b = append(b, k.Delete)
	}
	if k.Quit.Enabled() {
		b = append(b, k.Quit)
	}
	if k.Esc.Enabled() {
		b = append(b, k.Esc)
	}
	if k.Help.Enabled() {
		b = append(b, k.Help)
	}
	return b
}

func (k KeyMap) FullHelp() [][]key.Binding {
	b := []key.Binding{}
	if k.Up.Enabled() {
		b = append(b, k.Up)
	}
	if k.Down.Enabled() {
		b = append(b, k.Down)
	}
	if k.Enter.Enabled() {
		b = append(b, k.Enter)
	}
	if k.New.Enabled() {
		b = append(b, k.New)
	}
	if k.Edit.Enabled() {
		b = append(b, k.Edit)
	}
	if k.Delete.Enabled() {
		b = append(b, k.Delete)
	}
	if k.Quit.Enabled() {
		b = append(b, k.Quit)
	}
	if k.Esc.Enabled() {
		b = append(b, k.Esc)
	}
	if k.GlobalQuit.Enabled() {
		b = append(b, k.GlobalQuit)
	}
	if k.Help.Enabled() {
		b = append(b, k.Help)
	}
	return [][]key.Binding{b}
}
