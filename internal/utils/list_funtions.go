package utils

import "github.com/charmbracelet/bubbles/v2/list"

func ResetListFilterState(l *list.Model) {
	l.ResetFilter()
	l.SetFilterText("")
	l.SetFilterState(list.FilterState(list.Filtering))
}
