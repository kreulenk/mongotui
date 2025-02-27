package dbcoltable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// keyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type keyMap struct {
	Quit        key.Binding
	LineUp      key.Binding
	LineDown    key.Binding
	GotoTop     key.Binding
	GotoBottom  key.Binding
	Right       key.Binding
	Left        key.Binding
	Insert      key.Binding
	Enter       key.Binding
	Drop        key.Binding
	StartSearch key.Binding
	StopSearch  key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	LineUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	GotoTop: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to databaseStart"),
	),
	GotoBottom: key.NewBinding(
		key.WithKeys("databaseEnd", "G"),
		key.WithHelp("G/databaseEnd", "go to databaseEnd"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Insert: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "insert"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select collection"),
	),
	Drop: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "drop"),
	),
	StartSearch: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	StopSearch: key.NewBinding(
		key.WithKeys("esc", "/"),
		key.WithHelp("esc or '/'", "to exit filter"),
	),
}

func (m *Model) HelpView() string {
	if m.filterEnabled {
		return lipgloss.JoinHorizontal(lipgloss.Left, m.searchBar.View(), m.searchHelpView())
	}
	if m.cursorColumn == databasesColumn && m.databaseFilter != "" {
		return lipgloss.JoinHorizontal(lipgloss.Left, m.Help.View(keys), fmt.Sprintf(" (%s)", m.databaseFilter))
	} else if m.cursorColumn == collectionsColumn && m.collectionFilter != "" {
		return lipgloss.JoinHorizontal(lipgloss.Left, m.Help.View(keys), fmt.Sprintf(" (%s)", m.collectionFilter))
	}
	return m.Help.View(keys)
}

// searchHelpView is used when the user is in searchMode to filter the databases or collections
func (m *Model) searchHelpView() string {
	return m.Help.ShortHelpView([]key.Binding{keys.StopSearch})
}

// ShortHelp implements the keyMap interface.
func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Quit, km.LineUp, km.LineDown, km.Right, km.Left, km.Drop, km.Insert, km.StartSearch}
}

// FullHelp is required to satisfy the keyMap interface
func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		km.ShortHelp(),
	}
}
