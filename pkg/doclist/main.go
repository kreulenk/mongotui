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
	Help help.Model

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

// New creates a new baseModel for the table widget.
func New(engine *mongodata.Engine) Model {
	m := Model{
		docs:     []Row{},
		viewport: viewport.New(0, 20),

		Help:   help.New(),
		styles: defaultStyles(),

		engine: engine,
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
		case key.Matches(msg, keys.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, keys.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keys.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, keys.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, keys.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, keys.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, keys.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keys.GotoTop):
			m.GotoTop()
		case key.Matches(msg, keys.GotoBottom):
			m.GotoBottom()
			//case key.Matches(msg, keys.Edit):
			//	m.EditDoc()
			//case key.Matches(msg, keys.View):
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
	m.UpdateViewport()
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.cursor = renderutils.Clamp(m.cursor+n, 0, len(m.docs)-1)
	m.UpdateViewport()
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
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
