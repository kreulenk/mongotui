package dbcoltable

import "github.com/charmbracelet/bubbles/key"

// keyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type keyMap struct {
	Quit       key.Binding
	LineUp     key.Binding
	LineDown   key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding
	Right      key.Binding
	Left       key.Binding
	Enter      key.Binding
	Delete     key.Binding
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
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select collection"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
}

func (m *Model) HelpView() string {
	return m.Help.View(keys)
}

// ShortHelp implements the keyMap interface.
func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Quit, km.LineUp, km.LineDown, km.Right, km.Left, km.Delete}
}

// FullHelp is required to satisfy the keyMap interface
func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		km.ShortHelp(),
	}
}
