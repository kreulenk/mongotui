package dbcoltable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/dbcolsearch"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"github.com/kreulenk/mongotui/pkg/renderutils"
	"github.com/mattn/go-runewidth"
	"go.mongodb.org/mongo-driver/v2/bson"
	"os"
	"strings"
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

	searchBar        *dbcolsearch.Model
	searchEnabled    bool
	databaseSearch   string // Used to filter the database list
	collectionSearch string

	engine *mongoengine.Engine
}

// New creates a new baseModel for the dbcoltable component
func New(engine *mongoengine.Engine, state *state.MainViewState) *Model {
	if err := engine.RefreshDbAndCollections(); err != nil {
		fmt.Printf("could not initialize data: %v\n", err)
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

		searchBar: dbcolsearch.New(),

		engine: engine,
	}
	return &m
}

// Update is the Bubble Tea update loop.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchEnabled {
			return m, m.handleSearchUpdate(msg)
		}

		switch {
		case key.Matches(msg, keys.LineUp):
			return m, m.MoveUp(1)
		case key.Matches(msg, keys.LineDown):
			return m, m.MoveDown(1)
		case key.Matches(msg, keys.GotoTop):
			return m, m.GotoTop()
		case key.Matches(msg, keys.GotoBottom):
			return m, m.GotoBottom()
		case key.Matches(msg, keys.Right):
			return m, m.MoveRight()
		case key.Matches(msg, keys.Left):
			m.MoveLeft()
			return m, nil
		case key.Matches(msg, keys.Enter):
			if m.cursorColumn == collectionsColumn {
				m.blur()
			}
			return m, nil
		case key.Matches(msg, keys.Delete):
			if m.cursorColumn == databasesColumn {
				return m, modal.DisplayDatabaseDeleteModal(m.cursoredDatabase())
			} else {
				return m, modal.DisplayCollectionDeleteModal(m.cursoredDatabase(), m.cursoredCollection())
			}
		case key.Matches(msg, keys.StartSearch):
			if m.cursorColumn == databasesColumn {
				m.searchBar.SetValue(m.databaseSearch)
			} else {
				m.searchBar.SetValue(m.collectionSearch)
			}
			m.searchEnabled = true
		}
	case modal.ExecColDelete:
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		m.engine.SetSelectedCollection(msg.DbName, m.getFilteredCollections()[m.cursorCollection])
		if len(m.getFilteredCollections()) == 1 { // If we are about to drop last collection making db disappear
			m.cursorColumn = databasesColumn
		}
		return m, m.engine.DropCollection(msg.DbName, msg.CollectionName)
	case modal.ExecDbDelete:
		m.cursorDatabase = renderutils.Max(0, m.cursorDatabase-1)
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		if len(m.getFilteredDbs()) == 1 { // If we are about to drop last collection making db disappear
			m.engine.SetSelectedDatabase("")
		} else {
			m.engine.SetSelectedDatabase(m.getFilteredDbs()[m.cursorDatabase])
		}
		return m, m.engine.DropDatabase(msg.DbName)
	}
	return m, nil
}

func (m *Model) handleSearchUpdate(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, keys.StopSearch) {
		m.searchEnabled = false
	} else if m.cursorColumn == databasesColumn {
		m.searchBar.Update(msg)
		// Reset value and return if search is too specific
		if len(filterBySearch(m.engine.GetDatabases(), m.searchBar.GetValue())) == 0 {
			m.searchBar.SetValue(m.databaseSearch)
			return nil
		}
		m.databaseSearch = m.searchBar.GetValue()
		dbMaxIndex := len(m.getFilteredDbs()) - 1
		if m.cursorDatabase > dbMaxIndex {
			m.cursorDatabase = dbMaxIndex
		}
		m.engine.SetSelectedDatabase(m.cursoredDatabase())
	} else {
		originalCollection := m.cursoredCollection()
		m.searchBar.Update(msg)
		// Reset value and return if search is too specific
		if len(filterBySearch(m.engine.GetSelectedCollections(), m.searchBar.GetValue())) == 0 {
			m.searchBar.SetValue(m.collectionSearch)
			return nil
		}

		m.collectionSearch = m.searchBar.GetValue()
		colMaxIndex := len(m.getFilteredCollections()) - 1
		if m.cursorCollection > colMaxIndex {
			m.cursorCollection = colMaxIndex
		}

		if originalCollection != m.cursoredCollection() {
			m.engine.SetSelectedCollection(m.getFilteredDbs()[m.cursorDatabase], m.getFilteredCollections()[m.cursorCollection])
			return m.engine.QueryCollection(bson.D{})
		}
	}
	return nil
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

// updateViewport renders all of the cells for the databases and collections that are displayed within the dbcoltable
func (m *Model) updateViewport() {
	// Database column
	if m.cursorDatabase >= 0 {
		m.databaseStart = renderutils.Clamp(m.cursorDatabase-m.viewport.Height+1, 0, m.cursorDatabase)
	} else {
		m.databaseStart = 0
	}
	m.databaseEnd = renderutils.Clamp(m.cursorDatabase+m.viewport.Height, m.cursorDatabase, len(m.getFilteredDbs()))
	renderedDbCells := make([]string, 0, len(m.getFilteredDbs()))
	for i := m.databaseStart; i < m.databaseEnd; i++ {
		renderedDbCells = append(renderedDbCells, m.renderDatabaseCell(i))
	}
	renderedDbColumn := lipgloss.JoinVertical(lipgloss.Top, renderedDbCells...)

	// Collection column
	if m.cursorCollection >= 0 {
		m.collectionStart = renderutils.Clamp(m.cursorCollection-m.viewport.Height+1, 0, m.cursorCollection)
	} else {
		m.collectionStart = 0
	}
	m.collectionEnd = renderutils.Clamp(m.cursorCollection+m.viewport.Height, m.cursorCollection, len(m.getFilteredCollections()))
	renderedCollectionCells := make([]string, 0, len(m.getFilteredCollections()))
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
	if m.cursorDatabase < 0 || m.cursorDatabase >= len(m.getFilteredDbs()) {
		return ""
	}
	return m.getFilteredDbs()[m.cursorDatabase]
}

// cursoredCollection returns the collection that is currently highlighted.
func (m *Model) cursoredCollection() string {
	if m.cursorCollection < 0 ||
		m.cursorCollection >= len(m.getFilteredCollections()) ||
		m.cursorColumn == databasesColumn {
		return ""
	}

	return m.getFilteredCollections()[m.cursorCollection]
}

// SetWidth sets the width of the viewport of the dbcoltable.
func (m *Model) SetWidth(dbColWidth, fullTermWidth int) {
	m.viewport.Width = dbColWidth
	m.searchBar.SetWidth(fullTermWidth - 16) // filter help menu is 16 chars
}

// SetHeight sets the height of the viewport of the dbcoltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) tea.Cmd {
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase-n, 0, len(m.getFilteredDbs())-1)
		m.engine.SetSelectedDatabase(m.getFilteredDbs()[m.cursorDatabase])
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection-n, 0, len(m.getFilteredCollections())-1)
		m.engine.SetSelectedCollection(m.getFilteredDbs()[m.cursorDatabase], m.getFilteredCollections()[m.cursorCollection])
		return m.engine.QueryCollection(bson.D{})
	}
	return nil
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) tea.Cmd {
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase+n, 0, len(m.getFilteredDbs())-1)
		m.engine.SetSelectedDatabase(m.getFilteredDbs()[m.cursorDatabase])
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection+n, 0, len(m.getFilteredCollections())-1)
		m.engine.SetSelectedCollection(m.getFilteredDbs()[m.cursorDatabase], m.getFilteredCollections()[m.cursorCollection])
		return m.engine.QueryCollection(bson.D{})
	}
	return nil
}

// MoveRight moves the column to the right.
func (m *Model) MoveRight() tea.Cmd {
	if m.cursorColumn == collectionsColumn {
		m.blur()
	} else if m.cursorColumn == databasesColumn {
		m.cursorColumn = collectionsColumn
		m.cursorCollection = 0
		m.engine.SetSelectedCollection(m.getFilteredDbs()[m.cursorDatabase], m.getFilteredCollections()[m.cursorCollection])
		return m.engine.QueryCollection(bson.D{})
	}
	return nil
}

// MoveLeft moves the column to the left.
func (m *Model) MoveLeft() {
	if m.cursorColumn == collectionsColumn {
		m.cursorCollection = 0
		m.collectionSearch = ""
	}
	m.cursorColumn = databasesColumn
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() tea.Cmd {
	if m.cursorColumn == databasesColumn {
		return m.MoveUp(m.cursorDatabase)
	} else {
		return m.MoveUp(m.cursorCollection)
	}
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() tea.Cmd {
	if m.cursorColumn == databasesColumn {
		return m.MoveDown(len(m.getFilteredDbs()))
	} else {
		return m.MoveDown(len(m.getFilteredCollections()))
	}
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
	renderedCell := m.styles.Cell.Render(runewidth.Truncate(m.getFilteredDbs()[r], m.columnWidth(), "…"))
	if r == m.cursorDatabase {
		renderedCell = m.styles.Selected.Render(renderedCell)
	}

	return renderedCell
}

func (m *Model) renderCollectionCell(r int) string {
	m.styles.Cell = m.styles.Cell.Width(m.columnWidth()).MaxWidth(m.columnWidth())
	renderedCell := m.styles.Cell.Render(runewidth.Truncate(m.getFilteredCollections()[r], m.columnWidth(), "…"))
	if r == m.cursorCollection && m.cursorColumn == collectionsColumn {
		renderedCell = m.styles.Selected.Render(renderedCell)
	}
	return renderedCell
}

// getFilteredDbs gets the latest list of databases from the mongoengine and then applies the filter that the user
// has entered
func (m *Model) getFilteredDbs() []string {
	return filterBySearch(m.engine.GetDatabases(), m.databaseSearch)
}

// getFilteredCollections gets the latest list of collections from the mongoengine and then applies the filter that the user
// has entered
func (m *Model) getFilteredCollections() []string {
	return filterBySearch(m.engine.GetSelectedCollections(), m.collectionSearch)
}

// filterBySearch is used to filter what databases or collections are viewable or selectable based on the search query
func filterBySearch(strSlice []string, filter string) []string {
	if filter == "" {
		return strSlice
	}
	var filteredSlice []string
	for _, s := range strSlice {
		if strings.Contains(s, filter) {
			filteredSlice = append(filteredSlice, s)
		}
	}
	return filteredSlice
}
