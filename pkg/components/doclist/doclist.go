package doclist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/components/searchbar"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"github.com/kreulenk/mongotui/pkg/renderutils"
	"github.com/mattn/go-runewidth"
	"strings"
)

type Model struct {
	state *state.MainViewState
	Help  help.Model

	searchBar *searchbar.Model
	styles    Styles

	cursor   int
	viewport viewport.Model

	engine *mongoengine.Engine
}

// New creates a new baseModel for the dbcoltable widget.
func New(engine *mongoengine.Engine, state *state.MainViewState) *Model {
	m := Model{
		state:     state,
		Help:      help.New(),
		searchBar: searchbar.New(),
		viewport:  viewport.New(0, 20),

		styles: defaultStyles(),

		engine: engine,
	}
	return &m
}

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
		case key.Matches(msg, keys.Edit):
			m.EditDoc()
		case key.Matches(msg, keys.View):
			m.ViewDoc()
		case key.Matches(msg, keys.Delete):
			m.engine.SetSelectedDocument(m.engine.GetQueriedDocs()[m.cursor])
			return m, modal.DisplayDocDeleteModal(m.engine.GetSelectedDocument())
		}
	case modal.ExecDocDelete:
		m.cursor = renderutils.Max(0, m.cursor-1)
		return m, m.engine.DropDocument(msg.Doc)
	}

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// View renders the component.
func (m *Model) View() string {
	m.updateViewport()
	return m.styles.Table.Render(m.viewport.View())
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

func (m *Model) ResetSearchBar() {
	m.searchBar.ResetValue()
}

func (m *Model) Focus() {
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57"))
}

func (m *Model) blur() {
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
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
		if m.engine.Skip+mongoengine.Limit > m.engine.DocCount {
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

// SetWidth sets the width of the viewport of the dbcoltable.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.searchBar.SetWidth(w)
}

// SetHeight sets the height of the viewport of the dbcoltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h
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
	if m.cursor == docIndex {
		return m.styles.SelectedDoc.Width(m.viewport.Width - 2).Render(s), heightLeft
	}
	return m.styles.Doc.Width(m.viewport.Width - 2).Render(s), heightLeft
}
