package modal

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Tab      key.Binding
	LineUp   key.Binding
	LineDown key.Binding
	Right    key.Binding
	Left     key.Binding
	Enter    key.Binding
}

var keys = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "tab"),
	),
	LineUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	LineDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
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
		key.WithHelp("enter", "confirm"),
	),
}
