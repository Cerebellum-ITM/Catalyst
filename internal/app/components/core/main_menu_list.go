package core

import (
	"fmt"
	"io"

	"catalyst/internal/app/styles"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
)

type MainMenuDelegate struct {
	list.DefaultDelegate
	Theme styles.Theme
}

func (d MainMenuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	theme := d.Theme
	baseStyle := theme.AppStyles().Base
	item, ok := listItem.(MenuItem)
	if !ok {
		return
	}
	if index == m.Index() {
		cursor := baseStyle.Foreground(theme.Accent).Render("❯")
		renderedTitle := baseStyle.Foreground(theme.Primary).Render(item.title)
		fmt.Fprintf(w, "%s %s", cursor, renderedTitle)
	} else {
		renderedTitle := baseStyle.Foreground(d.Theme.Blur).Render(item.title)
		fmt.Fprintf(w, "  %s", renderedTitle)
	}
}

type MenuItem struct {
	title string
	value int
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Value() int          { return i.value }
func (i MenuItem) Description() string { return "" }
func (i MenuItem) FilterValue() string { return i.title }

func NewMainMenu(theme styles.Theme) list.Model {
	items := []list.Item{
		MenuItem{title: "Get Runes", value: 0},
		MenuItem{title: "Create Rune", value: 1},
		MenuItem{title: "Manage Loegs", value: 2},
		MenuItem{title: "View History", value: 3},
		MenuItem{title: "Demo Lock Screen", value: 4},
	}

	mainList := list.New(items, MainMenuDelegate{Theme: theme}, 0, 0)
	mainList.SetShowHelp(false)
	mainList.SetShowTitle(false)
	mainList.SetShowStatusBar(false)
	mainList.SetFilteringEnabled(true)
	mainList.KeyMap.AcceptWhileFiltering = key.NewBinding(
		key.WithKeys("enter", "/", "up", "down"),
	)
	mainList.KeyMap.CancelWhileFiltering = key.NewBinding(key.WithKeys("/"))
	return mainList
}
