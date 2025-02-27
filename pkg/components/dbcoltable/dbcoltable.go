package dbcoltable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
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

	viewport viewport.Model

	cursorColumn     cursorColumn
	cursorDatabase   int
	cursorCollection int

	databaseStart   int
	databaseEnd     int
	collectionStart int
	collectionEnd   int

	searchBar        textinput.Model
	filterEnabled    bool
	databaseFilter   string // Used to filter the database list
	collectionFilter string

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

	ti := textinput.New()
	ti.Placeholder = "Filter"
	ti.Focus()

	m := Model{
		state:  state,
		Help:   help.New(),
		styles: defaultStyles(),

		viewport: viewport.New(0, 20),

		cursorColumn:   databasesColumn,
		cursorDatabase: 0,

		searchBar: ti,

		engine: engine,
	}
	return &m
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

// getFilteredDbs gets the latest list of databases from the mongoengine and then applies the filter that the user
// has entered
func (m *Model) getFilteredDbs() []string {
	return filterBySearch(m.engine.GetDatabases(), m.databaseFilter)
}

// getFilteredCollections gets the latest list of collections from the mongoengine and then applies the filter that the user
// has entered
func (m *Model) getFilteredCollections() []string {
	return filterBySearch(m.engine.GetSelectedCollections(), m.collectionFilter)
}

// filterBySearch is used to filter what databases or collections are viewable or selectable based on the search query
func filterBySearch(strSlice []string, filter string) []string {
	if filter == "" {
		return strSlice
	}
	var filteredSlice []string
	for _, s := range strSlice {
		if strings.Contains(strings.ToLower(s), strings.ToLower(filter)) {
			filteredSlice = append(filteredSlice, s)
		}
	}
	return filteredSlice
}

func (m *Model) IsFilterEnabled() bool {
	return m.filterEnabled
}
