package dbcoltable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/internal/state"
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
	state  *state.TuiState
	Help   help.Model
	styles Styles

	viewport    viewport.Model
	headersText []string
	databases   []string
	collections []string

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
func New(engine *mongoengine.Engine, state *state.TuiState) *Model {
	err := engine.RefreshDatabases()
	if err != nil {
		fmt.Printf("could not initialize data: %v", err)
		os.Exit(1)
	}

	m := Model{
		state:  state,
		Help:   help.New(),
		styles: defaultStyles(),

		viewport:    viewport.New(0, 20),
		headersText: []string{"Databases", "Collections"},
		databases:   mongoengine.GetSortedDatabasesByName(engine.Server.Databases),
		collections: []string{},

		cursorColumn: databasesColumn,

		engine: engine,
	}

	err = m.updateCollectionsData()
	if err != nil {
		fmt.Printf("failed to initialize the database and collection information: %v\n", err)
		os.Exit(1)
	}
	m.updateViewport()

	return &m
}

// Update is the Bubble Tea update loop.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.LineUp):
			m.MoveUp(1)
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
				m.blur()
			}
		case key.Matches(msg, keys.Delete):
			m.DeleteDbOrCol()
		}
	}

	return m, nil
}

func (m *Model) RefreshAfterDeletion() {
	dbColState := m.state.MainViewState.DbColTableState
	if dbColState.WasDatabaseDeletedViaModal() { // Selectively reset cursors
		m.cursorDatabase = renderutils.Max(0, m.cursorDatabase-1)
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		m.state.MainViewState.DbColTableState.ResetDatabaseDeletionRefreshFlag()
		m.RefreshData()
	} else if dbColState.WasCollectionDeletedViaModal() {
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		m.state.MainViewState.DbColTableState.ResetCollectionDeletionRefreshFlag()
		m.RefreshData()
	}
}

func (m *Model) RefreshData() {
	m.engine.ClearCachedData()
	m.state.MainViewState.DbColTableState.ClearCollectionSelection()
	if err := m.refreshDatabasesData(); err != nil {
		m.state.ModalState.SetError(fmt.Errorf("unable to reinitialize databases after deletion: %w", err))
	}
	if err := m.updateCollectionsData(); err != nil {
		m.state.ModalState.SetError(fmt.Errorf("unable to reinitialize collections after deletion: %w", err))
	}
	m.updateViewport()
}

// Focus enables key use on the dbcoltable so that the user can navigate the dbcoltable again. This signal would
// be sent from another component
func (m *Model) Focus() {
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57"))
	m.updateViewport()
}

// blur disables key use on the dbcoltable so that the parent mainview component can switch the focus to
// the doclist component
func (m *Model) blur() {
	m.state.MainViewState.SetActiveComponent(state.DocList)
	m.state.MainViewState.DbColTableState.SetSelectedCollection(m.cursoredDatabase(), m.cursoredCollection())
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
}

func (m *Model) DeleteDbOrCol() {
	m.state.MainViewState.DbColTableState.SetSelectedCollection(m.cursoredDatabase(), m.cursoredCollection())
	if m.cursorColumn == databasesColumn {
		m.state.ModalState.RequestDatabaseModalDeletionPrompt()
	} else {
		m.state.ModalState.RequestCollectionModalDeletionPrompt()
	}
}

// View renders the component.
func (m *Model) View() string {
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

// cursoredDatabase returns the database that is currently highlighted.
func (m *Model) cursoredDatabase() string {
	if m.cursorDatabase < 0 || m.cursorDatabase >= len(m.databases) {
		return ""
	}
	return m.databases[m.cursorDatabase]
}

// cursoredCollection returns the collection that is currently highlighted.
func (m *Model) cursoredCollection() string {
	if m.cursorCollection < 0 ||
		m.cursorCollection >= len(m.collections) ||
		m.cursorColumn == databasesColumn {
		return ""
	}

	return m.collections[m.cursorCollection]
}

// SetWidth sets the width of the viewport of the dbcoltable.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.updateViewport()
}

// SetHeight sets the height of the viewport of the dbcoltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
	m.updateViewport()
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	oldCursor := m.cursorDatabase
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase-n, 0, len(m.databases)-1)
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection-n, 0, len(m.collections)-1)
	}

	err := m.updateCollectionsData()
	if err != nil {
		m.state.ModalState.SetError(err)
		m.cursorDatabase = oldCursor
		return
	}
	m.updateViewport()
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	oldCursor := m.cursorDatabase
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase+n, 0, len(m.databases)-1)
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection+n, 0, len(m.collections)-1)
	}

	err := m.updateCollectionsData()
	if err != nil {
		m.state.ModalState.SetError(err)
		m.cursorDatabase = oldCursor
		return
	}
	m.updateViewport()
}

// MoveRight moves the column to the right.
func (m *Model) MoveRight() {
	if m.cursorColumn == collectionsColumn {
		m.blur()
		return
	} else if m.cursorColumn == databasesColumn {
		m.cursorColumn = collectionsColumn
		err := m.updateCollectionsData()
		if err != nil {
			m.state.ModalState.SetError(err)
			m.cursorColumn = databasesColumn
			return
		}
		m.updateViewport()
	}
}

// MoveLeft moves the column to the left.
func (m *Model) MoveLeft() {
	oldCursorCollection := m.cursorCollection
	oldCursorColumn := m.cursorColumn
	if m.cursorColumn == collectionsColumn {
		m.cursorCollection = 0
	}
	m.cursorColumn = databasesColumn
	err := m.updateCollectionsData()
	if err != nil {
		m.state.ModalState.SetError(err)
		m.cursorCollection = oldCursorCollection
		m.cursorColumn = oldCursorColumn
		return
	}
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
	renderedCell := m.styles.Cell.Render(runewidth.Truncate(m.databases[r], m.columnWidth(), "…"))
	if r == m.cursorDatabase {
		renderedCell = m.styles.Selected.Render(renderedCell)
	}

	return renderedCell
}

func (m *Model) renderCollectionCell(r int) string {
	m.styles.Cell = m.styles.Cell.Width(m.columnWidth()).MaxWidth(m.columnWidth())
	renderedCell := m.styles.Cell.Render(runewidth.Truncate(m.collections[r], m.columnWidth(), "…"))
	if r == m.cursorCollection && m.cursorColumn == collectionsColumn {
		renderedCell = m.styles.Selected.Render(renderedCell)
	}

	return renderedCell
}

func (m *Model) refreshDatabasesData() error {
	err := m.engine.RefreshDatabases()
	if err != nil {
		return fmt.Errorf("failed to load databases")
	}
	m.databases = mongoengine.GetSortedDatabasesByName(m.engine.Server.Databases)
	return nil
}

// updateCollectionsData updates the data tracked in the model based on the current cursorDatabase, cursorCollection and cursorColumn position
// Lots of opportunity for caching with how this function is handled/called, but I like the live data for now
func (m *Model) updateCollectionsData() error {
	collections, err := m.engine.GetCollectionsOfDb(m.cursoredDatabase())
	if err != nil {
		return err
	}
	m.collections = collections
	return nil
}
