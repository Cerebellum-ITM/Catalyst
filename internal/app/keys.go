package app

import "github.com/charmbracelet/bubbles/v2/key"

// KeyMap defines a set of keybindings.
// It implements the help.KeyMap interface.
type KeyMap struct {
	// General
	SwitchFocus key.Binding
	Enter       key.Binding
	Quit        key.Binding
	GlobalQuit  key.Binding
	Help        key.Binding
	Esc         key.Binding

	// Navigation
	Up     key.Binding
	Down   key.Binding
	PgUp   key.Binding
	PgDown key.Binding

	// Rune specific
	Edit          key.Binding
	Delete        key.Binding
	New           key.Binding
	ClearFilter   key.Binding
	NextField     key.Binding
	AddCommand    key.Binding
	RemoveCommand key.Binding
	submit        key.Binding
}

func viewPortKeys() KeyMap {
	return KeyMap{
		PgUp:        key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PgDown:      key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", "page down")),
		SwitchFocus: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Toggle focus")),
		Quit:        key.NewBinding(key.WithKeys("q"), key.WithHelp("q/ctrl+x", "quit")),
		GlobalQuit:  key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func mainListKeys() KeyMap {
	return KeyMap{
		Up:          key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:        key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Quit:        key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		SwitchFocus: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Toggle focus")),
		GlobalQuit:  key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		ClearFilter: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "clear filter"),
		),
	}
}

func viewingRunesKeys() KeyMap {
	return KeyMap{
		Up:          key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:        key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "run")),
		Edit:        key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "edit")),
		Delete:      key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
		SwitchFocus: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Toggle focus")),
		Esc:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		GlobalQuit:  key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		ClearFilter: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "clear filter"),
		),
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
		Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "enter")),
		submit: key.NewBinding(
			key.WithKeys("shift+enter"),
			key.WithHelp("shift+enter", "submit form"),
		),
		AddCommand: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "Add cmd"),
		),
		RemoveCommand: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "Remove cmd"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "previous field"),
		),
		Down:        key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "next field")),
		Esc:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		SwitchFocus: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Toggle focus")),
		GlobalQuit:  key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func executingRuneKeys() KeyMap {
	return KeyMap{
		Up:          key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
		Down:        key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
		SwitchFocus: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch focus")),
		Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "back")),
		Esc:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		GlobalQuit:  key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "quit")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
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
	if k.AddCommand.Enabled() {
		b = append(b, k.AddCommand)
	}
	if k.RemoveCommand.Enabled() {
		b = append(b, k.RemoveCommand)
	}
	if k.Enter.Enabled() {
		b = append(b, k.Enter)
	}
	if k.SwitchFocus.Enabled() {
		b = append(b, k.SwitchFocus)
	}
	if k.PgUp.Enabled() {
		b = append(b, k.PgUp)
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
	if k.PgUp.Enabled() {
		b = append(b, k.PgUp)
	}
	if k.PgDown.Enabled() {
		b = append(b, k.PgDown)
	}
	if k.Up.Enabled() {
		b = append(b, k.Up)
	}
	if k.Down.Enabled() {
		b = append(b, k.Down)
	}
	if k.Enter.Enabled() {
		b = append(b, k.Enter)
	}
	if k.SwitchFocus.Enabled() {
		b = append(b, k.SwitchFocus)
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
	if k.ClearFilter.Enabled() {
		b = append(b, k.ClearFilter)
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
