package doclist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/errormodal"
	"github.com/kreulenk/mongotui/pkg/components/searchbar"
	"github.com/kreulenk/mongotui/pkg/mongodata"
	"github.com/kreulenk/mongotui/pkg/renderutils"
	"github.com/mattn/go-runewidth"
	"go.mongodb.org/mongo-driver/v2/bson"
	"os"
	"slices"
	"strings"
)

type DocAction int

const (
	NoAction DocAction = iota
	ViewDoc
	EditDoc
)

type Model struct {
	Help     help.Model
	errModal *errormodal.Model

	searchBar     *searchbar.Model
	searchEnabled bool

	docs   []Doc
	styles Styles

	cursor    int
	viewport  viewport.Model
	focus     bool
	docAction DocAction

	engine *mongodata.Engine
}

type Doc []FieldSummary

type FieldSummary struct {
	Name  string
	Type  string // TODO restrict to a set of types
	Value string
}

// New creates a new baseModel for the dbcoltable widget.
func New(engine *mongodata.Engine, errModal *errormodal.Model) *Model {
	m := Model{
		Help:      help.New(),
		errModal:  errModal,
		searchBar: searchbar.New(errModal),

		docs:     []Doc{},
		viewport: viewport.New(0, 20),

		styles: defaultStyles(),

		docAction: NoAction,
		engine:    engine,
	}

	err := m.updateTableRows()
	if err != nil {
		fmt.Printf("could not initialize document summary list: %v", err)
		os.Exit(1)
	}
	m.updateViewport()

	return &m
}

// Update is the Bubble Tea update loop.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	if m.searchEnabled {
		m.searchBar, _ = m.searchBar.Update(msg)
		if !m.searchBar.Focused() {
			m.searchEnabled = false
			err := m.updateTableRows()
			if err != nil {
				m.errModal.SetError(err)
				return m, nil
			}
		}
		m.updateViewport()
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Esc):
			m.blur()
		case key.Matches(msg, keys.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, keys.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keys.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, keys.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, keys.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, keys.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, keys.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, keys.GotoTop):
			m.GotoTop()
		case key.Matches(msg, keys.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, keys.Left):
			m.engine.ClearCollectionSelection()
			m.blur()
		case key.Matches(msg, keys.Edit):
			m.EditDoc()
		case key.Matches(msg, keys.View):
			m.ViewDoc()
		}
	}

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// View renders the component.
func (m *Model) View() string {
	return m.styles.Table.Render(m.viewport.View())
}

// updateTableRows updates the list of document summaries shown in the right hand bar based on the database/collection
// selected in the left hand bar as well as the query that was last entered in the search bar.
func (m *Model) updateTableRows() error {
	if !m.engine.IsCollectionSelected() {
		return nil
	}
	val, err := m.searchBar.GetValue()
	if err != nil {
		return err
	}
	err = m.engine.QueryCollection(val)
	if err != nil {
		return err
	}

	var newDocs []Doc
	for _, doc := range m.engine.GetQueriedDocs() {
		var row Doc
		for k, v := range doc {
			val := fmt.Sprintf("%v", v)
			fType := getFieldType(v)
			if fType == "Object" || fType == "Array" { // TODO restrict to a set of types
				val = fType
			}
			row = append(row, FieldSummary{
				Name:  k,
				Type:  fType,
				Value: val,
			})
		}
		newDocs = append(newDocs, row)
	}
	m.cursor = 0
	m.docs = newDocs
	return nil
}

func getFieldType(value interface{}) string {
	switch v := value.(type) {
	case bson.M, bson.D:
		return "Object"
	case bson.A:
		return "Array"
	case nil:
		return "Null"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus(resetValues bool) {
	m.docAction = NoAction
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57"))
	if resetValues {
		m.searchBar.ResetValue()
		err := m.updateTableRows()
		if err != nil {
			m.errModal.SetError(err)
		}
		m.updateViewport()
	}
	m.focus = true
}

func (m *Model) blur() {
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	m.focus = false
}

func (m *Model) GetDocAction() DocAction {
	return m.docAction
}

func (m *Model) EditDoc() {
	if m.cursor >= len(m.docs) {
		m.errModal.SetError(fmt.Errorf("no document selected"))
		return
	}
	
	m.docAction = EditDoc
	m.engine.SetSelectedDocument(m.cursor)
	m.blur()
}

// ViewDoc opens the selected document in a new window via the jsonviewer component.
func (m *Model) ViewDoc() {
	if m.cursor >= len(m.docs) {
		m.errModal.SetError(fmt.Errorf("no document selected"))
		return
	}

	m.docAction = ViewDoc
	m.engine.SetSelectedDocument(m.cursor)
	m.blur()
}

func (m *Model) IsDocSelected() bool {
	return m.cursor >= 0 && m.cursor < len(m.docs)
}

// updateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) updateViewport() {
	renderedRows := make([]string, 0, len(m.docs))
	var startDocIndex = m.getStartIndex()
	heightLeft := m.viewport.Height
	for i := startDocIndex; i < len(m.docs) && heightLeft >= 3; i++ { // 3 as one cell is minimum 3 lines
		newRow, heightUsed := m.renderDocSummary(i, heightLeft)
		heightLeft -= heightUsed
		renderedRows = append(renderedRows, newRow)
	}

	joinedRows := lipgloss.JoinVertical(lipgloss.Top, renderedRows...)
	if len(m.docs) == 0 && m.engine.IsCollectionSelected() {
		joinedRows = "\nNo documents found" // TODO make this centered
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Top, m.searchBar.View(), joinedRows),
	)
}

// getStartIndex returns the index of the first row to be displayed in the viewport.
// It is calculated based on the cursor position and the size of each row cell based on if there are 4 or less fields
// per doc
func (m *Model) getStartIndex() int {
	if len(m.docs) == 0 {
		return 0
	}
	heightLeft := m.viewport.Height - 3 // 1 to account for the search bar and top/bottom borders
	startIndex := m.cursor
	for i := m.cursor; i >= 0 && heightLeft > 0; i-- {
		heightLeft -= 2 // To account for the space between rows from borders
		if len(m.docs[i]) > 4 {
			heightLeft -= 4
		} else {
			heightLeft -= len(m.docs[i])
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
	m.updateViewport()
}

// SetHeight sets the height of the viewport of the dbcoltable.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h
	m.updateViewport()
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	if m.cursor == 0 {
		m.searchEnabled = true
		m.searchBar.Focus()
	}

	m.cursor = renderutils.Clamp(m.cursor-n, 0, len(m.docs)-1)
	m.updateViewport()
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.cursor = renderutils.Clamp(m.cursor+n, 0, len(m.docs)-1)
	m.updateViewport()
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	m.MoveDown(len(m.docs))
}

func (m *Model) renderDocSummary(docIndex, heightLeft int) (string, int) {
	heightLeft -= 2 // To account for the space between rows
	doc := m.docs[docIndex]
	slices.SortFunc(doc, func(i, j FieldSummary) int {
		return strings.Compare(i.Name, j.Name)
	})
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
