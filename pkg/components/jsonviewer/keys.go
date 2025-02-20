package jsonviewer

import "github.com/charmbracelet/bubbles/key"

// keyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type keyMap struct {
	Back     key.Binding
	LineUp   key.Binding
	LineDown key.Binding
}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.LineUp, km.LineDown, km.Back}
}

// FullHelp is only used to satisfy the interface as we do not actually use this
func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		km.ShortHelp(),
	}
}

var keys = keyMap{
	Back: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "back"),
	),
	LineUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
}
