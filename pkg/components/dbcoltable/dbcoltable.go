package dbcoltable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/internal/state"
	"github.com/kreulenk/mongotui/pkg/components/modal"
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
	err := engine.RefreshDbAndCollections()
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
		collections: []string{},

		cursorColumn:   databasesColumn,
		cursorDatabase: 0,

		engine: engine,
	}
	m.updateViewport()

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
			//m.DeleteDbOrCol()
		}
	}
	if err != nil {
		return m, modal.DisplayErrorModal(err)
	}

	return m, nil
}

func (m *Model) RefreshAfterDeletion() error {
	dbColState := m.state.MainViewState.DbColTableState
	if dbColState.WasDatabaseDeletedViaModal() { // Selectively reset cursors
		m.cursorDatabase = renderutils.Max(0, m.cursorDatabase-1)
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		m.state.MainViewState.DbColTableState.ResetDatabaseDeletionRefreshFlag()
		return m.RefreshData()
	} else if dbColState.WasCollectionDeletedViaModal() {
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		m.state.MainViewState.DbColTableState.ResetCollectionDeletionRefreshFlag()
		return m.RefreshData()
	}
	return nil
}

func (m *Model) RefreshData() error {
	m.engine.ClearCachedData()
	m.state.MainViewState.DbColTableState.ClearCollectionSelection()
	if err := m.engine.RefreshDbAndCollections(); err != nil {
		return fmt.Errorf("unable to reinitialize databases after deletion: %w", err)
	}
	if err := m.updateCollectionsData(); err != nil {
		return fmt.Errorf("unable to reinitialize collections after deletion: %w", err)
	}
	m.updateViewport()
	return nil
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

//func (m *Model) DeleteDbOrCol() {
//	m.state.MainViewState.DbColTableState.SetSelectedCollection(m.cursoredDatabase(), m.cursoredCollection())
//	if m.cursorColumn == databasesColumn {
//		m.state.ModalState.RequestDatabaseModalDeletionPrompt()
//	} else {
//		m.state.ModalState.RequestCollectionModalDeletionPrompt()
//	}
//}

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
	if m.cursorDatabase < 0 || m.cursorDatabase >= len(m.engine.GetDatabases()) {
		return ""
	}
	return m.engine.GetDatabaseByIndex(m.cursorDatabase)
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
func (m *Model) MoveUp(n int) error {
	oldCursorDb := m.cursorDatabase
	oldCursorCol := m.cursorCollection
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase-n, 0, len(m.engine.GetDatabases())-1)
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection-n, 0, len(m.collections)-1)
	}

	err := m.updateCollectionsData()
	if err != nil {
		m.cursorDatabase = oldCursorDb
		m.cursorCollection = oldCursorCol
		return err
	}
	m.updateViewport()
	return nil
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) error {
	oldCursorDb := m.cursorDatabase
	oldCursorCol := m.cursorCollection
	if m.cursorColumn == databasesColumn {
		m.cursorDatabase = renderutils.Clamp(m.cursorDatabase+n, 0, len(m.engine.GetDatabases())-1)
	} else {
		m.cursorCollection = renderutils.Clamp(m.cursorCollection+n, 0, len(m.collections)-1)
	}

	err := m.updateCollectionsData()
	if err != nil {
		m.cursorDatabase = oldCursorDb
		m.cursorCollection = oldCursorCol
		return err
	}
	m.updateViewport()
	return nil
}

// MoveRight moves the column to the right.
func (m *Model) MoveRight() error {
	if m.cursorColumn == collectionsColumn {
		m.blur()
		return nil
	} else if m.cursorColumn == databasesColumn {
		m.cursorColumn = collectionsColumn
		err := m.updateCollectionsData()
		if err != nil {
			m.cursorColumn = databasesColumn
			return err
		}
		m.updateViewport()
	}
	return nil
}

// MoveLeft moves the column to the left.
func (m *Model) MoveLeft() error {
	oldCursorCollection := m.cursorCollection
	oldCursorColumn := m.cursorColumn
	if m.cursorColumn == collectionsColumn {
		m.cursorCollection = 0
	}
	m.cursorColumn = databasesColumn
	err := m.updateCollectionsData()
	if err != nil {
		m.cursorCollection = oldCursorCollection
		m.cursorColumn = oldCursorColumn
		return err
	}
	m.updateViewport()
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
		return m.MoveDown(len(m.collections))
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
	renderedCell := m.styles.Cell.Render(runewidth.Truncate(m.engine.GetDatabaseByIndex(r), m.columnWidth(), "…"))
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

// updateCollectionsData updates the data tracked in the model based on the current cursorDatabase, cursorCollection and cursorColumn position
// Lots of opportunity for caching with how this function is handled/called, but I like the live data for now
func (m *Model) updateCollectionsData() error {
	collections, err := m.engine.FetchCollectionsPerDb(m.cursoredDatabase())
	if err != nil {
		return err
	}
	m.collections = collections
	return nil
}
