package table

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
	"os"
)

type cursorColumn int

const (
	databasesColumn cursorColumn = iota
	collectionsColumn
)

// Model defines a state for the table widget.
type Model struct {
	Help help.Model

	viewport    viewport.Model
	headersText []string
	databases   []string
	collections []string

	cursorColumn     cursorColumn
	cursorDatabase   int
	cursorCollection int
	focus            bool

	databaseStart   int
	databaseEnd     int
	collectionStart int
	collectionEnd   int

	engine *mongodata.Engine
}

// New creates a new baseModel for the table widget.
func New(engine *mongodata.Engine) Model {
	databases := mongodata.GetSortedDatabasesByName(engine.Server.Databases)
	m := Model{
		Help: help.New(),

		viewport:    viewport.New(0, 20),
		headersText: []string{"Databases", "Collections"},
		databases:   databases,
		collections: []string{},

		cursorColumn: databasesColumn,
		focus:        true,

		engine: engine,
	}

	m.updateTableData()
	m.UpdateViewport()

	return m
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, keyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keyMap.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, keyMap.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, keyMap.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, keyMap.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, keyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, keyMap.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, keyMap.Right):
			m.GoRight()
		case key.Matches(msg, keyMap.Left):
			m.GoLeft()
		}
	}

	return m, nil
}

// Focused returns the focus state of the table.
func (m Model) Focused() bool {
	return m.focus
}

// Focus focuses the table, allowing the user to move around the rows and
// interact.
func (m *Model) Focus() {
	m.focus = true
	m.UpdateViewport()
}

// Blur blurs the table, preventing selection or movement.
func (m *Model) Blur() {
	m.focus = false
	m.UpdateViewport()
}

// View renders the component.
func (m Model) View() string {
	return styles.Table.Render(m.headersView() + "\n" + m.viewport.View())
}

// UpdateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) UpdateViewport() {
	// Database column
	if m.cursorDatabase >= 0 {
		m.databaseStart = renderutils.Clamp(m.cursorDatabase-m.viewport.Height, 0, m.cursorDatabase)
	} else {
		m.databaseStart = 0
	}
	m.databaseEnd = renderutils.Clamp(m.cursorDatabase+m.viewport.Height, m.cursorDatabase, len(m.databases))
	renderedDbCells := make([]string, 0, len(m.databases))
	for i := m.databaseStart; i < m.databaseEnd; i++ {
		renderedDbCells = append(renderedDbCells, m.renderDatabaseCell(i))
	}
	renderedDbColumn := lipgloss.JoinVertical(lipgloss.Top, renderedDbCells...)

	// Collection column
	if m.cursorCollection >= 0 {
		m.collectionStart = renderutils.Clamp(m.cursorCollection-m.viewport.Height, 0, m.cursorCollection)
	} else {
		m.collectionStart = 0
	}
	m.collectionEnd = renderutils.Clamp(m.cursorCollection+m.viewport.Height, m.cursorCollection, len(m.collections))
	renderedCollectionCells := make([]string, 0, len(m.collections))
	for i := m.collectionStart; i < m.collectionEnd; i++ {
		renderedCollectionCells = append(renderedCollectionCells, m.renderCollectionCell(i))
	}
	renderedCollectionColumn := lipgloss.JoinVertical(lipgloss.Top, renderedCollectionCells...)

	m.viewport.SetContent(
		lipgloss.JoinHorizontal(lipgloss.Left, renderedDbColumn, renderedCollectionColumn),
	)
}

// SelectedRow returns the selected row.
func (m Model) SelectedDatabase() string {
	if m.cursorDatabase < 0 || m.cursorDatabase >= len(m.databases) {
		return ""
	}

	return m.databases[m.cursorDatabase]
}

// SelectedCell returns the text within the currently highlighted cell
func (m Model) SelectedCell() string {
	if m.cursorDatabase < 0 || m.cursorDatabase >= len(m.databases) {
		return ""
	} else if m.cursorCollection < 0 || m.cursorCollection >= len(m.collections) {
		return ""
	}

	if m.cursorColumn == databasesColumn {
		return m.databases[m.cursorDatabase]
	} else if m.cursorColumn == collectionsColumn {
		return m.collections[m.cursorCollection]
	}
	return ""
}

// SetWidth sets the width of the viewport of the table.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.UpdateViewport()
}

// SetHeight sets the height of the viewport of the table.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
	m.UpdateViewport()
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase-n, 0, len(m.databases)-1)
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection-n, 0, len(m.collections)-1)
	}

	m.updateTableData()
	m.UpdateViewport()
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase+n, 0, len(m.databases)-1)
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection+n, 0, len(m.collections)-1)
	}

	m.updateTableData()
	m.UpdateViewport()
}

// MoveRight moves the column to the right.
func (m *Model) MoveRight() {
	m.cursorColumn = collectionsColumn
	m.updateTableData()
	m.UpdateViewport()
}

// MoveLeft moves the column to the left.
func (m *Model) MoveLeft() {
	if m.cursorColumn == collectionsColumn {
		m.cursorCollection = 0
	}
	m.cursorColumn = databasesColumn
	m.updateTableData()
	m.UpdateViewport()
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	if m.cursorColumn == databasesColumn {
		m.MoveUp(m.cursorDatabase)
	} else {
		m.MoveUp(m.cursorCollection)
	}
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	if m.cursorColumn == databasesColumn {
		m.MoveDown(len(m.databases))
	} else {
		m.MoveDown(len(m.collections))
	}
}

// GoRight moves to the next column.
func (m *Model) GoRight() {
	m.MoveRight()
}

func (m *Model) GoLeft() {
	m.MoveLeft()
}

func (m *Model) columnWidth() int {
	return m.viewport.Width / len(m.headersText)
}
func (m Model) headersView() string {
	s := make([]string, 0, len(m.headersText))
	for _, col := range m.headersText {
		style := lipgloss.NewStyle().Width(m.columnWidth()).MaxWidth(m.columnWidth()).Inline(true)
		renderedCell := style.Render(runewidth.Truncate(col, m.viewport.Width/2, "…"))
		s = append(s, styles.Header.Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, s...)
}

func (m *Model) renderDatabaseCell(r int) string {
	style := lipgloss.NewStyle().Width(m.columnWidth()).MaxWidth(m.columnWidth()).Inline(true)
	renderedCell := styles.Cell.Render(style.Render(runewidth.Truncate(m.databases[r], m.viewport.Width/len(m.headersText), "…")))
	if r == m.cursorDatabase {
		renderedCell = styles.Selected.Render(renderedCell)
	}

	return renderedCell
}

func (m *Model) renderCollectionCell(r int) string {
	style := lipgloss.NewStyle().Width(m.columnWidth()).MaxWidth(m.columnWidth()).Inline(true)
	renderedCell := styles.Cell.Render(style.Render(runewidth.Truncate(m.collections[r], m.viewport.Width/len(m.headersText), "…")))
	if r == m.cursorCollection && m.cursorColumn == collectionsColumn {
		renderedCell = styles.Selected.Render(renderedCell)
	}

	return renderedCell
}

// updateTableData updates the data tracked in the model based on the current cursorDatabase, cursorCollection and cursorColumn position
// Lots of opportunity for caching with how this function is handled/called, but I like the live data for now
func (m *Model) updateTableData() {
	err := m.engine.SetCollectionsPerDb(m.databases[m.cursorDatabase])
	if err != nil { // TODO handle errors better
		fmt.Printf("could not fetch collections: %v", err)
		os.Exit(1)
	}

	m.collections = mongodata.GetSortedCollectionsByName(m.engine.Server.Databases[m.databases[m.cursorDatabase]].Collections)
}
