package dbcoltable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"github.com/kreulenk/mongotui/pkg/renderutils"
	"github.com/mattn/go-runewidth"
	"os"
)

type cursorColumn int

const (
	databasesColumn cursorColumn = iota
	collectionsColumn
)

// Model defines a state for the dbcoltable widget.
type Model struct {
	state  *state.MainViewState
	Help   help.Model
	styles Styles

	viewport    viewport.Model
	headersText []string

	cursorColumn     cursorColumn
	cursorDatabase   int
	cursorCollection int

	databaseStart   int
	databaseEnd     int
	collectionStart int
	collectionEnd   int

	engine *mongoengine.Engine
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// New creates a new baseModel for the dbcoltable widget.
func New(engine *mongoengine.Engine, state *state.MainViewState) *Model {
	err := engine.RefreshDbAndCollections()
	if err != nil {
		fmt.Printf("could not initialize data: %v\n", err()) // Crash and run tea.Cmd func to display err
		os.Exit(1)
	}
	if len(engine.GetDatabases()) > 0 {
		engine.SetSelectedDatabase(engine.GetDatabases()[0])
	}

	m := Model{
		state:  state,
		Help:   help.New(),
		styles: defaultStyles(),

		viewport:    viewport.New(0, 20),
		headersText: []string{"Databases", "Collections"},

		cursorColumn:   databasesColumn,
		cursorDatabase: 0,

		engine: engine,
	}
	return &m
}

// Update is the Bubble Tea update loop.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var err error
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.LineUp):
			err = m.MoveUp(1)
		case key.Matches(msg, keys.LineDown):
			err = m.MoveDown(1)
		case key.Matches(msg, keys.GotoTop):
			err = m.GotoTop()
		case key.Matches(msg, keys.GotoBottom):
			err = m.GotoBottom()
		case key.Matches(msg, keys.Right):
			err = m.GoRight()
		case key.Matches(msg, keys.Left):
			err = m.GoLeft()
		case key.Matches(msg, keys.Enter):
			if m.cursorColumn == collectionsColumn {
				m.blur()
			}
		case key.Matches(msg, keys.Delete):
			if m.cursorColumn == databasesColumn {
				return m, modal.DisplayDatabaseDeleteModal(m.cursoredDatabase())
			} else {
				return m, modal.DisplayCollectionDeleteModal(m.cursoredDatabase(), m.cursoredCollection())
			}
		}
	case modal.ExecColDelete:
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		m.engine.SetSelectedCollection(msg.DbName, m.engine.GetSelectedCollections()[m.cursorCollection])
		return m, m.engine.DropCollection(msg.DbName, msg.CollectionName)
	case modal.ExecDbDelete:
		m.cursorDatabase = renderutils.Max(0, m.cursorDatabase-1)
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		return m, m.engine.DropDatabase(msg.DbName)
	}
	if err != nil {
		return m, modal.DisplayErrorModal(err)
	}

	return m, nil
}

// Focus enables key use on the dbcoltable so that the user can navigate the dbcoltable again. This signal would
// be sent from another component
func (m *Model) Focus() {
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57"))
}

// blur disables key use on the dbcoltable so that the parent mainview component can switch the focus to
// the doclist component
func (m *Model) blur() {
	m.state.SetActiveComponent(state.DocList)
	m.engine.SetSelectedCollection(m.cursoredDatabase(), m.cursoredCollection())
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
}

// View renders the component.
func (m *Model) View() string {
	m.updateViewport()
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
	m.databaseEnd = renderutils.Clamp(m.cursorDatabase+m.viewport.Height, m.cursorDatabase, len(m.engine.GetDatabases()))
	renderedDbCells := make([]string, 0, len(m.engine.GetDatabases()))
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
	m.collectionEnd = renderutils.Clamp(m.cursorCollection+m.viewport.Height, m.cursorCollection, len(m.engine.GetSelectedCollections()))
	renderedCollectionCells := make([]string, 0, len(m.engine.GetSelectedCollections()))
	for i := m.collectionStart; i < m.collectionEnd; i++ {
		renderedCollectionCells = append(renderedCollectionCells, m.renderCollectionCell(i))
	}
	renderedCollectionColumn := lipgloss.JoinVertical(lipgloss.Top, renderedCollectionCells...)

	m.viewport.SetContent(
		lipgloss.JoinHorizontal(lipgloss.Left, renderedDbColumn, renderedCollectionColumn),
	)
}

// cursoredDatabase returns the database that is currently highlighted.
func (m *Model) cursoredDatabase() string {
	if m.cursorDatabase < 0 || m.cursorDatabase >= len(m.engine.GetDatabases()) {
		return ""
	}
	return m.engine.GetDatabases()[m.cursorDatabase]
}

// cursoredCollection returns the collection that is currently highlighted.
func (m *Model) cursoredCollection() string {
	if m.cursorCollection < 0 ||
		m.cursorCollection >= len(m.engine.GetSelectedCollections()) ||
		m.cursorColumn == databasesColumn {
		return ""
	}

	return m.engine.GetSelectedCollections()[m.cursorCollection]
}

// SetWidth sets the width of the viewport of the dbcoltable.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
}

// SetHeight sets the height of the viewport of the dbcoltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) error {
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase-n, 0, len(m.engine.GetDatabases())-1)
		m.engine.SetSelectedDatabase(m.engine.GetDatabases()[m.cursorDatabase])
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection-n, 0, len(m.engine.GetSelectedCollections())-1)
		m.engine.SetSelectedCollection(m.engine.GetDatabases()[m.cursorDatabase], m.engine.GetSelectedCollections()[m.cursorCollection])
	}

	return nil
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) error {
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase+n, 0, len(m.engine.GetDatabases())-1)
		m.engine.SetSelectedDatabase(m.engine.GetDatabases()[m.cursorDatabase])
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection+n, 0, len(m.engine.GetSelectedCollections())-1)
		m.engine.SetSelectedCollection(m.engine.GetDatabases()[m.cursorDatabase], m.engine.GetSelectedCollections()[m.cursorCollection])
	}
	return nil
}

// MoveRight moves the column to the right.
func (m *Model) MoveRight() error {
	if m.cursorColumn == collectionsColumn {
		m.blur()
		return nil
	} else if m.cursorColumn == databasesColumn {
		m.cursorColumn = collectionsColumn
		m.cursorCollection = 0
		m.engine.SetSelectedCollection(m.engine.GetDatabases()[m.cursorDatabase], m.engine.GetSelectedCollections()[m.cursorCollection])
	}
	return nil
}

// MoveLeft moves the column to the left.
func (m *Model) MoveLeft() error {
	if m.cursorColumn == collectionsColumn {
		m.cursorCollection = 0
	}
	m.cursorColumn = databasesColumn
	return nil
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() error {
	if m.cursorColumn == databasesColumn {
		return m.MoveUp(m.cursorDatabase)
	} else {
		return m.MoveUp(m.cursorCollection)
	}
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() error {
	if m.cursorColumn == databasesColumn {
		return m.MoveDown(len(m.engine.GetDatabases()))
	} else {
		return m.MoveDown(len(m.engine.GetSelectedCollections()))
	}
}

// GoRight moves to the next column.
func (m *Model) GoRight() error {
	return m.MoveRight()
}

func (m *Model) GoLeft() error {
	return m.MoveLeft()
}

func (m *Model) columnWidth() int {
	return m.viewport.Width / len(m.headersText)
}
func (m *Model) headersView() string {
	s := make([]string, 0, len(m.headersText))
	for _, col := range m.headersText {
		m.styles.Header = m.styles.Header.Width(m.columnWidth()).MaxWidth(m.columnWidth())
		renderedCell := m.styles.Header.Render(runewidth.Truncate(col, m.columnWidth(), "…"))
		s = append(s, renderedCell)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, s...)
}

func (m *Model) renderDatabaseCell(r int) string {
	m.styles.Cell = m.styles.Cell.Width(m.columnWidth()).MaxWidth(m.columnWidth())
	renderedCell := m.styles.Cell.Render(runewidth.Truncate(m.engine.GetDatabases()[r], m.columnWidth(), "…"))
	if r == m.cursorDatabase {
		renderedCell = m.styles.Selected.Render(renderedCell)
	}

	return renderedCell
}

func (m *Model) renderCollectionCell(r int) string {
	m.styles.Cell = m.styles.Cell.Width(m.columnWidth()).MaxWidth(m.columnWidth())
	renderedCell := m.styles.Cell.Render(runewidth.Truncate(m.engine.GetSelectedCollections()[r], m.columnWidth(), "…"))
	if r == m.cursorCollection && m.cursorColumn == collectionsColumn {
		renderedCell = m.styles.Selected.Render(renderedCell)
	}

	return renderedCell
}
