package doclist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"mtui/pkg/mongodata"
	"mtui/pkg/renderutils"
)

type Model struct {
	KeyMap KeyMap
	Help   help.Model

	docs   []Row
	styles Styles
	cursor int

	viewport viewport.Model
	start    int
	end      int

	engine *mongodata.Engine
}

type Row []DocSummary

type DocSummary struct {
	FieldName  string
	FieldType  string // TODO restrict to a set of types
	FieldValue string
}

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	Edit         key.Binding
	View         key.Binding
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.LineUp, km.LineDown, km.Edit, km.View}
}

// FullHelp implements the KeyMap interface.
func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Edit, km.View},
		{km.LineUp, km.LineDown, km.GotoTop, km.GotoBottom},
		{km.PageUp, km.PageDown, km.HalfPageUp, km.HalfPageDown},
	}
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	const spacebar = " "
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", spacebar),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit document"),
		),
		View: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view document"),
		),
	}
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Cell     lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns a set of default style definitions for this table.
func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

// SetStyles sets the table styles.
func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.UpdateViewport()
}

// Option is used to set options in New. For example:
//
//	table := New(WithColumns([]Column{{Title: "ID", Width: 10}}))
type Option func(*Model)

// New creates a new baseModel for the table widget.
func New(engine *mongodata.Engine, opts ...Option) Model {
	m := Model{
		docs:     []Row{},
		viewport: viewport.New(0, 20),

		KeyMap: DefaultKeyMap(),
		Help:   help.New(),
		styles: DefaultStyles(),

		engine: engine,
	}

	for _, opt := range opts {
		opt(&m)
	}

	//m.updateTableRows() // TODO add so that we actually have some data to display
	m.UpdateViewport()

	return m
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
			//case key.Matches(msg, m.KeyMap.Edit):
			//	m.EditDoc()
			//case key.Matches(msg, m.KeyMap.View):
			//	m.ViewDoc()
		}
	}

	return m, nil
}

// View renders the component.
func (m Model) View() string {
	return "TODO: implement doc view"
	//return m.viewport.View()
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m Model) HelpView() string {
	return m.Help.View(m.KeyMap)
}

// UpdateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) UpdateViewport() {
	renderedRows := make([]string, 0, len(m.docs))

	// Render only rows from: m.cursorRow-m.viewport.Height to: m.cursorRow+m.viewport.Height
	// Constant runtime, independent of number of rows in a table.
	// Limits the number of renderedRows to a maximum of 2*m.viewport.Height
	if m.cursor >= 0 {
		m.start = renderutils.Clamp(m.cursor-m.viewport.Height, 0, m.cursor)
	} else {
		m.start = 0
	}
	m.end = renderutils.Clamp(m.cursor+m.viewport.Height, m.cursor, len(m.docs))
	for i := m.start; i < m.end; i++ {
		renderedRows = append(renderedRows, m.renderRow(i))
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

// SetWidth sets the width of the viewport of the table.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.UpdateViewport()
}

// SetHeight sets the height of the viewport of the table.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h
	m.UpdateViewport()
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	m.cursor = renderutils.Clamp(m.cursor-n, 0, len(m.docs)-1)

	// TODO investigate if we need this switch since there's no headers on this element
	switch {
	case m.start == 0:
		m.viewport.SetYOffset(renderutils.Clamp(m.viewport.YOffset, 0, m.cursor))
	case m.start < m.viewport.Height:
		m.viewport.YOffset = renderutils.Clamp(renderutils.Clamp(m.viewport.YOffset+n, 0, m.cursor), 0, m.viewport.Height)
	case m.viewport.YOffset >= 1:
		m.viewport.YOffset = renderutils.Clamp(m.viewport.YOffset+n, 1, m.viewport.Height)
	}
	m.UpdateViewport()
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.cursor = renderutils.Clamp(m.cursor+n, 0, len(m.docs)-1)
	m.UpdateViewport()

	// TODO investigate if we need this switch since there's no headers on this element
	switch {
	case m.end == len(m.docs) && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(renderutils.Clamp(m.viewport.YOffset-n, 1, m.viewport.Height))
	case m.cursor > (m.end-m.start)/2 && m.viewport.YOffset > 0:
		m.viewport.SetYOffset(renderutils.Clamp(m.viewport.YOffset-n, 1, m.cursor))
	case m.viewport.YOffset > 1:
	case m.cursor > m.viewport.YOffset+m.viewport.Height-1:
		m.viewport.SetYOffset(renderutils.Clamp(m.viewport.YOffset+1, 0, 1))
	}
}

// GotoTop moves the selection to the first row.
// TODO fix this to work with this new split table element
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
// TODO fix this to work with this new split table element
func (m *Model) GotoBottom() {
	m.MoveDown(len(m.docs))
}

func (m *Model) renderRow(r int) string {
	s := make([]string, 0, len(m.docs))

	rowText := fmt.Sprintf("%v", r) // TODO actually render the text of the row. Maybe the first four fields using a template?

	style := lipgloss.NewStyle().Width(m.viewport.Width).MaxWidth(m.viewport.Width).Inline(true)
	renderedCell := m.styles.Cell.Render(style.Render(runewidth.Truncate(rowText, m.viewport.Width, "…")))
	s = append(s, renderedCell)

	row := lipgloss.JoinHorizontal(lipgloss.Top, s...)
	return row
}
