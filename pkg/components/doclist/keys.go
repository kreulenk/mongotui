package doclist

import "github.com/charmbracelet/bubbles/key"

// keyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type keyMap struct {
	Quit       key.Binding
	Esc        key.Binding
	LineUp     key.Binding
	LineDown   key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding
	NextPage   key.Binding // pagination
	PrevPage   key.Binding // pagination
	Left       key.Binding
	Insert     key.Binding
	Edit       key.Binding
	View       key.Binding
	Delete     key.Binding
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m *Model) HelpView() string {
	return m.Help.View(keys)
}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Quit, km.LineUp, km.LineDown, km.PrevPage, km.NextPage, km.Insert, km.Edit, km.View, km.Delete}
}

// FullHelp is needed to satisfy the keyMap interface
func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		km.ShortHelp(),
	}
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
		key.WithHelp("g/home", "go to start"),
	),
	GotoBottom: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "next page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "previous page"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Insert: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "insert"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	View: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "view"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
}
