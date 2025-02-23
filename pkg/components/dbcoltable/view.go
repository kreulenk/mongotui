package dbcoltable

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/renderutils"
	"github.com/mattn/go-runewidth"
)

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

// SetWidth sets the width of the viewport of the dbcoltable.
func (m *Model) SetWidth(dbColWidth, fullTermWidth int) {
	m.viewport.Width = dbColWidth
	m.searchBar.SetWidth(fullTermWidth - 23) // filter help menu is 23 chars
}

// SetHeight sets the height of the viewport of the dbcoltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h - lipgloss.Height(m.headersView())
}

func (m *Model) columnWidth() int {
	return m.viewport.Width / 2
}

func (m *Model) headersView() string {
	m.styles.Header = m.styles.Header.Width(m.columnWidth()).MaxWidth(m.columnWidth())
	dbText := fmt.Sprintf("Databases (%d)", len(m.getFilteredDbs()))
	collectionText := fmt.Sprintf("Collections (%d)", len(m.getFilteredCollections()))

	dbCell := m.styles.Header.Render(runewidth.Truncate(dbText, m.columnWidth(), "…"))
	collectionCell := m.styles.Header.Render(runewidth.Truncate(collectionText, m.columnWidth(), "…"))
	return lipgloss.JoinHorizontal(lipgloss.Top, dbCell, collectionCell)
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
