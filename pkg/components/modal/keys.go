package modal

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Right key.Binding
	Left  key.Binding
	Enter key.Binding
}

var keys = keyMap{
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
