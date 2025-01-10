package doclist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mtui/pkg/mongodata"
	"mtui/pkg/renderutils"
	"os"
	"slices"
	"strings"
)

type selectedCollection struct {
	collectionName string
	databaseName   string
}

type Model struct {
	Help help.Model

	docs   []Doc
	styles Styles

	cursor             int
	selectedCollection selectedCollection

	viewport viewport.Model
	focus    bool

	engine *mongodata.Engine
}

type Doc []FieldSummary

type FieldSummary struct {
	Name  string
	Type  string // TODO restrict to a set of types
	Value string
}

// New creates a new baseModel for the table widget.
func New(engine *mongodata.Engine) Model {
	m := Model{
		docs:     []Doc{},
		viewport: viewport.New(0, 20),

		Help:   help.New(),
		styles: defaultStyles(),

		engine: engine,
	}

	m.updateTableRows()
	m.updateViewport()

	return m
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
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
			m.blur()
			//case key.Matches(msg, keys.Edit):
			//	m.EditDoc()
			//case key.Matches(msg, keys.View):
			//	m.ViewDoc()
		}
	}

	return m, nil
}

// SetSelectedCollection Allows parent components to set what data will be displayed within this component.
func (m *Model) SetSelectedCollection(collectionName, databaseName string) {
	m.selectedCollection = selectedCollection{
		collectionName: collectionName,
		databaseName:   databaseName,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// View renders the component.
func (m Model) View() string {
	return m.styles.Table.Render(m.viewport.View())
}

func (m *Model) updateTableRows() {
	mongoDocs, err := m.engine.GetData(m.selectedCollection.databaseName, m.selectedCollection.collectionName)
	if err != nil { // TODO improve how we handle errors
		fmt.Printf("could not get data: %v", err)
		os.Exit(1)
	}
	var newDocs []Doc
	for _, doc := range mongoDocs {
		var row Doc
		for k, v := range doc {
			row = append(row, FieldSummary{
				Name:  k,
				Type:  getFieldType(v),
				Value: fmt.Sprintf("%v", v),
			})
		}
		newDocs = append(newDocs, row)
	}
	m.docs = newDocs
}

// TODO add the remaining data types that mongo supports
func getFieldType(value interface{}) string {
	switch v := value.(type) {
	case string, int32, int64, float64, bool, primitive.ObjectID, primitive.Timestamp, primitive.DateTime:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		return "Object"
	case []interface{}:
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

func (m *Model) Focus() {
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57"))
	m.updateTableRows()
	m.updateViewport()
	m.focus = true
}

func (m *Model) blur() {
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	m.focus = false
}

// updateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) updateViewport() {
	renderedRows := make([]string, 0, len(m.docs))
	heightLeft := m.viewport.Height
	for i := m.cursor; i < len(m.docs) && heightLeft > 0; i++ {
		newRow, heightUsed := m.renderDocSummary(i, heightLeft)
		renderedRows = append(renderedRows, newRow)
		heightLeft -= heightUsed
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

// SetWidth sets the width of the viewport of the table.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.updateViewport()
}

// SetHeight sets the height of the viewport of the table.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h
	m.updateViewport()
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
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
		if i == 4 || heightLeft <= 0 { // Only show the first 4 fields and make sure we have not exceeded viewport height
			break
		}
		newField := fmt.Sprintf("%s: %s", m.styles.DocText.Render(field.Name), field.Value)
		fields = append(fields, newField)
		heightLeft--
	}

	s := lipgloss.JoinVertical(lipgloss.Top, fields...)
	if m.cursor == docIndex {
		return m.styles.SelectedDoc.Width(m.viewport.Width).MaxWidth(m.viewport.Width).Render(s), heightLeft
	} else {
		return m.styles.Doc.Width(m.viewport.Width).MaxWidth(m.viewport.Width).Render(s), heightLeft
	}
}
