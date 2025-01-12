package coltable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mtui/pkg/mongodata"
	"github.com/kreulenk/mtui/pkg/renderutils"
	"github.com/mattn/go-runewidth"
	"os"
)

type cursorColumn int

const (
	databasesColumn cursorColumn = iota
	collectionsColumn
)

// Model defines a state for the coltable widget.
type Model struct {
	Help   help.Model
	styles Styles

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

// New creates a new baseModel for the coltable widget.
func New(engine *mongodata.Engine) Model {
	databases := mongodata.GetSortedDatabasesByName(engine.Server.Databases)
	m := Model{
		Help:   help.New(),
		styles: defaultStyles(),

		viewport:    viewport.New(0, 20),
		headersText: []string{"Databases", "Collections"},
		databases:   databases,
		collections: []string{},

		cursorColumn: databasesColumn,
		focus:        true,

		engine: engine,
	}

	m.updateTableData()
	m.updateViewport()

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
		case key.Matches(msg, keys.Right):
			m.GoRight()
		case key.Matches(msg, keys.Left):
			m.GoLeft()
		case key.Matches(msg, keys.Enter):
			if m.cursorColumn == collectionsColumn {
				m.selectCollection()
			}
		}
	}

	return m, nil
}

// CollectionSelected returns whether or not a collection is currently selected.
func (m Model) CollectionSelected() bool {
	return !m.focus
}

// DeselectCollection enables key use on the coltable so that the user can navigate the coltable again. This signal would
// be sent from another component
func (m *Model) DeselectCollection() {
	m.focus = true
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57"))
	m.updateViewport()
}

// selectCollection disables key use on the coltable so that other components can now display information about the collection
// that is currently highlighted
func (m *Model) selectCollection() {
	m.focus = false
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	m.updateViewport()
}

// View renders the component.
func (m Model) View() string {
	return m.styles.Table.Render(m.headersView() + "\n" + m.viewport.View())
}

// updateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) updateViewport() {
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

// SelectedDatabase returns the database that is currently highlighted.
func (m Model) SelectedDatabase() string {
	if m.cursorDatabase < 0 || m.cursorDatabase >= len(m.databases) {
		return ""
	}
	return m.databases[m.cursorDatabase]
}

// SelectedCollection returns the collection that is currently highlighted.
func (m Model) SelectedCollection() string {
	if m.cursorCollection < 0 ||
		m.cursorCollection >= len(m.collections) ||
		m.cursorColumn == databasesColumn {
		return ""
	}

	return m.collections[m.cursorCollection]
}

// SetWidth sets the width of the viewport of the coltable.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.updateViewport()
}

// SetHeight sets the height of the viewport of the coltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
	m.updateViewport()
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
	m.updateViewport()
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
	m.updateViewport()
}

// MoveRight moves the column to the right.
func (m *Model) MoveRight() {
	if m.cursorColumn == collectionsColumn {
		m.selectCollection()
		return
	} else if m.cursorColumn == databasesColumn {
		m.cursorColumn = collectionsColumn
		m.updateTableData()
		m.updateViewport()
	}
}

// MoveLeft moves the column to the left.
func (m *Model) MoveLeft() {
	if m.cursorColumn == collectionsColumn {
		m.cursorCollection = 0
	}
	m.cursorColumn = databasesColumn
	m.updateTableData()
	m.updateViewport()
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
		s = append(s, m.styles.Header.Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, s...)
}

func (m *Model) renderDatabaseCell(r int) string {
	style := lipgloss.NewStyle().Width(m.columnWidth()).MaxWidth(m.columnWidth()).Inline(true)
	renderedCell := m.styles.Cell.Render(style.Render(runewidth.Truncate(m.databases[r], m.viewport.Width/len(m.headersText), "…")))
	if r == m.cursorDatabase {
		renderedCell = m.styles.Selected.Render(renderedCell)
	}

	return renderedCell
}

func (m *Model) renderCollectionCell(r int) string {
	style := lipgloss.NewStyle().Width(m.columnWidth()).MaxWidth(m.columnWidth()).Inline(true)
	renderedCell := m.styles.Cell.Render(style.Render(runewidth.Truncate(m.collections[r], m.viewport.Width/len(m.headersText), "…")))
	if r == m.cursorCollection && m.cursorColumn == collectionsColumn {
		renderedCell = m.styles.Selected.Render(renderedCell)
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
