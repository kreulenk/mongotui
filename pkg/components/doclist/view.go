package doclist

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"github.com/mattn/go-runewidth"
)

// View renders the component.
func (m *Model) View() string {
	m.updateViewport()
	return m.styles.Table.Render(m.viewport.View())
}

// SetWidth sets the width of the viewport of the dbcoltable.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.searchBar.SetWidth(w)
}

// SetHeight sets the height of the viewport of the dbcoltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h
}

func (m *Model) IsDocSelected() bool {
	return m.cursor >= 0 && m.cursor < len(m.engine.GetDocumentSummaries())
}

// updateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) updateViewport() {
	renderedRows := make([]string, 0, len(m.engine.GetDocumentSummaries()))
	var startDocIndex = m.getStartIndex()
	heightLeft := m.viewport.Height - 2 // 2 to account for search bar and pagination info

	for i := startDocIndex; i < len(m.engine.GetDocumentSummaries()) && heightLeft >= 3; i++ { // 3 as one cell is minimum 3 lines
		newRow, heightUsed := m.renderDocSummary(i, heightLeft)
		heightLeft -= heightUsed
		renderedRows = append(renderedRows, newRow)
	}
	joinedRows := lipgloss.JoinVertical(lipgloss.Top, renderedRows...)
	if len(m.engine.GetDocumentSummaries()) == 0 {
		joinedRows = "\nNo documents found" // TODO make this centered
	}

	if m.engine.DocCount > 0 {
		var paginationTracker string
		if m.engine.DocCount == 1 {
			paginationTracker = "viewing document 1 of 1"
		} else if m.engine.Skip+mongoengine.Limit > m.engine.DocCount {
			paginationTracker = fmt.Sprintf("viewing documents %d-%d of %d", m.engine.Skip+1, m.engine.DocCount, m.engine.DocCount)
		} else {
			paginationTracker = fmt.Sprintf("viewing documents %d-%d of %d", m.engine.Skip+1, m.engine.Skip+mongoengine.Limit, m.engine.DocCount)
		}
		m.viewport.SetContent(
			lipgloss.JoinVertical(lipgloss.Top, m.searchBar.View(), lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Right, paginationTracker), joinedRows))
	} else {
		m.viewport.SetContent(
			lipgloss.JoinVertical(lipgloss.Top, m.searchBar.View(), joinedRows),
		)
	}
}

func (m *Model) renderDocSummary(docIndex, heightLeft int) (string, int) {
	heightLeft -= 2 // To account for the space between rows
	doc := m.engine.GetDocumentSummaries()[docIndex]
	var fields []string
	for i, field := range doc {
		if i >= 4 || heightLeft < 0 { // Only show the first 4 fields and make sure we have not exceeded viewport height
			break
		}
		// Colors are not properly counted in runewidth so we have to do calculations before applying any styling
		if runewidth.StringWidth(field.Name) > m.viewport.Width-2 {
			fieldName := runewidth.Truncate(field.Name, m.viewport.Width-2, "…")
			fields = append(fields, m.styles.DocText.Render(fieldName))
		} else {
			fieldValue := runewidth.Truncate(field.Value, m.viewport.Width-4-runewidth.StringWidth(field.Name), "…")
			newField := fmt.Sprintf("%s: %s", m.styles.DocText.Render(field.Name), fieldValue)
			fields = append(fields, newField)
		}

		heightLeft--
	}

	s := lipgloss.JoinVertical(lipgloss.Top, fields...)
	if m.cursor == docIndex && m.focused && !m.searchBar.Focused() {
		return m.styles.SelectedDoc.Width(m.viewport.Width - 2).Render(s), heightLeft
	}
	return m.styles.Doc.Width(m.viewport.Width - 2).Render(s), heightLeft
}

// getStartIndex returns the index of the first row to be displayed in the viewport.
// It is calculated based on the cursor position and the size of each row cell based on if there are 4 or less fields
// per doc
func (m *Model) getStartIndex() int {
	if len(m.engine.GetDocumentSummaries()) == 0 {
		return 0
	}
	heightLeft := m.viewport.Height - 3 // 1 to account for the search bar and top/bottom borders
	startIndex := m.cursor
	for i := m.cursor; i >= 0 && heightLeft > 0; i-- {
		heightLeft -= 2 // To account for the space between rows from borders
		if len(m.engine.GetDocumentSummaries()[i]) > 4 {
			heightLeft -= 4
		} else {
			heightLeft -= len(m.engine.GetDocumentSummaries()[i])
		}
		if heightLeft >= 0 {
			startIndex = i
		}
	}
	return startIndex
}
