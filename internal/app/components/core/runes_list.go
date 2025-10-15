package core

import (
	"fmt"
	"io"

	"catalyst/internal/app/styles"
	"catalyst/internal/types"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
)

type RunesListDelegate struct {
	list.DefaultDelegate
	Theme styles.Theme
}

func (d RunesListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	theme := d.Theme
	baseStyle := theme.AppStyles().Base
	item, ok := listItem.(RuneItem)
	if !ok {
		return
	}

	var queue string
	if item.QueuePosition > 0 {
		queue = fmt.Sprintf("[%d]", item.QueuePosition)
	} else {
		queue = ""
	}

	if index == m.Index() {
		cursor := baseStyle.Foreground(theme.Accent).Render("❯")
		renderedTitle := baseStyle.Foreground(theme.Primary).Render(item.Title())
		rendererQueue := baseStyle.Foreground(theme.Secondary).Render(queue)
		fmt.Fprintf(w, "%s %s %s", cursor, renderedTitle, rendererQueue)
	} else {
		renderedTitle := baseStyle.Foreground(d.Theme.Blur).Render(item.Title())
		rendererQueue := baseStyle.Foreground(theme.Secondary).Render(queue)
		fmt.Fprintf(w, "  %s %s", renderedTitle, rendererQueue)
	}
}

type RuneItem struct {
	types.Rune
	QueuePosition int
}

func (i RuneItem) Title() string       { return i.Rune.Name }
func (i RuneItem) Description() string { return i.Rune.Description }
func (i RuneItem) FilterValue() string { return i.Rune.Name }

func NewRunesList(theme styles.Theme, runes []types.Rune) list.Model {
	items := make([]list.Item, len(runes))
	for i, r := range runes {
		items[i] = RuneItem{Rune: r}
	}

	runesList := list.New(items, RunesListDelegate{Theme: theme}, 0, 0)
	runesList.SetShowHelp(false)
	runesList.SetShowTitle(false)
	runesList.SetShowStatusBar(false)
	runesList.SetFilteringEnabled(true)
	runesList.KeyMap.AcceptWhileFiltering = key.NewBinding(
		key.WithKeys("enter", "/", "up", "down"),
	)
	runesList.KeyMap.CancelWhileFiltering = key.NewBinding(key.WithKeys("/"))

	return runesList
}
