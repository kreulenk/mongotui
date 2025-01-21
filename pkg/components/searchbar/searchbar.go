package searchbar

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Model struct {
	textInput textinput.Model
	textValue string

	msgModal *modal.Model // TODO look if we actually want to pass msgMoal into this component
}

func New(msgModal *modal.Model) *Model {
	ti := textinput.New()
	ti.Placeholder = "Query"
	ti.SetValue("{}")
	ti.SetCursor(1)
	ti.CharLimit = 156
	ti.Blur()

	return &Model{
		textValue: "{}",
		textInput: ti,
		msgModal:  msgModal,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

// SetWidth sets the width of the viewport of the dbcoltable.
func (m *Model) SetWidth(w int) {
	m.textInput.Width = w
}

func (m *Model) Focus() {
	m.textInput.Focus()
}

func (m *Model) Focused() bool {
	return m.textInput.Focused()
}

func (m *Model) ResetValue() {
	m.textInput.Reset()
	m.textValue = "{}"
	m.textInput.SetValue(m.textValue)
	m.textInput.SetCursor(1)
}

func (m *Model) GetValue() (bson.D, error) {
	var query bson.D
	err := bson.UnmarshalExtJSON([]byte(m.textValue), false, &query)
	if err != nil {
		return bson.D{}, fmt.Errorf("invalid query: %v", err)
	}
	return query, nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.textValue = m.textInput.Value()
			m.textInput.Blur()
			return m, nil
		case tea.KeyEsc, tea.KeyDown:
			m.textInput.Blur()
			return m, nil
		default:
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.textInput.View()
}
