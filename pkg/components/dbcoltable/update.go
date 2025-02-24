package dbcoltable

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/renderutils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Update is the Bubble Tea update loop.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterEnabled {
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
		case key.Matches(msg, keys.Drop):
			if m.cursorColumn == databasesColumn {
				return m, modal.DisplayDatabaseDropModal(m.cursoredDatabase())
			} else {
				return m, modal.DisplayCollectionDropModal(m.cursoredDatabase(), m.cursoredCollection())
			}
		case key.Matches(msg, keys.StartSearch):
			if m.cursorColumn == databasesColumn {
				m.searchBar.SetValue(m.databaseFilter)
			} else {
				m.searchBar.SetValue(m.collectionFilter)
			}
			m.filterEnabled = true
		}
	case modal.ExecCollDrop:
		m.cursorCollection = renderutils.Max(0, m.cursorCollection-1)
		m.engine.SetSelectedCollection(msg.DbName, m.getFilteredCollections()[m.cursorCollection])
		if len(m.getFilteredCollections()) == 1 { // If we are about to drop last collection making db disappear
			m.cursorColumn = databasesColumn
		}
		return m, m.engine.DropCollection(msg.DbName, msg.CollectionName)
	case modal.ExecDbDrop:
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
		m.filterEnabled = false
	} else if m.cursorColumn == databasesColumn {
		m.searchBar.Update(msg)
		// Reset value and return if search is too specific
		if len(filterBySearch(m.engine.GetDatabases(), m.searchBar.GetValue())) == 0 {
			m.searchBar.SetValue(m.databaseFilter)
			return nil
		}
		m.databaseFilter = m.searchBar.GetValue()
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
			m.searchBar.SetValue(m.collectionFilter)
			return nil
		}

		m.collectionFilter = m.searchBar.GetValue()
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
		m.collectionFilter = ""
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
