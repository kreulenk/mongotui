package doclist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/renderutils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
)

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	// Handle searchBar updates
	if m.searchBar.Focused() {
		if k, ok := msg.(tea.KeyMsg); ok { // If the user hit enter into the search bar, update the docList
			if strings.Trim(k.String(), " ") == "enter" { // For some reason enter always comes in with spaces
				m.searchBar.Blur()
				return m, m.ExecuteQuery()
			}
		}
		m.searchBar, _ = m.searchBar.Update(msg)
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, keys.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keys.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keys.GotoTop):
			m.GotoTop()
		case key.Matches(msg, keys.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, keys.NextPage):
			m.cursor = 0
			return m, m.engine.NextPage()
		case key.Matches(msg, keys.PrevPage):
			m.cursor = 0
			return m, m.engine.PreviousPage()
		case key.Matches(msg, keys.Left):
			m.state.SetActiveComponent(state.DbColTable)
			m.blur()
			m.searchBar.ResetValue()
			return m, m.engine.QueryCollection(bson.D{})
		case key.Matches(msg, keys.Insert):
			m.state.SetActiveComponent(state.DocInsert)
		case key.Matches(msg, keys.Edit):
			if len(m.engine.GetDocumentSummaries()) > 0 {
				m.EditDoc()
			} else {
				return m, modal.DisplayErrorModal(fmt.Errorf("cannot edit a document as none is selected"))
			}
		case key.Matches(msg, keys.View):
			if len(m.engine.GetDocumentSummaries()) > 0 {
				m.ViewDoc()
			} else {
				return m, modal.DisplayErrorModal(fmt.Errorf("cannot view a document as none is selected"))
			}
		case key.Matches(msg, keys.Delete):
			m.engine.SetSelectedDocument(m.engine.GetQueriedDocs()[m.cursor])
			return m, modal.DisplayDocDeleteModal(m.engine.GetSelectedDocument())
		}
	case modal.ExecDocDelete:
		m.cursor = renderutils.Max(0, m.cursor-1)
		return m, m.engine.DeleteDocument(msg.Doc)
	}

	return m, nil
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	if m.cursor == 0 {
		m.searchBar.Focus()
	}

	m.cursor = renderutils.Clamp(m.cursor-n, 0, len(m.engine.GetDocumentSummaries())-1)
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.cursor = renderutils.Clamp(m.cursor+n, 0, len(m.engine.GetDocumentSummaries())-1)
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	m.MoveDown(len(m.engine.GetDocumentSummaries()))
}

// ExecuteQuery updates the list of document summaries shown in the right hand bar based on the database/collection
// selected in the left hand bar as well as the query that was last entered in the search bar.
func (m *Model) ExecuteQuery() tea.Cmd {
	val, err := m.searchBar.GetValue()
	if err != nil {
		return modal.DisplayErrorModal(err)
	}
	m.cursor = 0
	return m.engine.QueryCollection(val)
}

func (m *Model) EditDoc() {
	m.state.SetActiveComponent(state.SingleDocEditor)
	m.engine.SetSelectedDocument(m.engine.GetQueriedDocs()[m.cursor])
}

// ViewDoc opens the selected document in a new window via the jsonviewer component.
func (m *Model) ViewDoc() {
	m.state.SetActiveComponent(state.SingleDocViewer)
	m.engine.SetSelectedDocument(m.engine.GetQueriedDocs()[m.cursor])
}
