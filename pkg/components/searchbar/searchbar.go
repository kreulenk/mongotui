package searchbar

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/internal/state"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Model struct {
	state     *state.TuiState
	textInput textinput.Model
}

func New(state *state.TuiState) *Model {
	ti := textinput.New()
	ti.Placeholder = "Query"
	ti.SetValue("{}")
	ti.SetCursor(1)
	ti.CharLimit = 156
	ti.Blur()

	return &Model{
		state:     state,
		textInput: ti,
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

func (m *Model) Blur() {
	m.textInput.Blur()
}

func (m *Model) Focused() bool {
	return m.textInput.Focused()
}

func (m *Model) ResetValue() {
	m.textInput.Reset()
	m.textInput.SetValue("{}")
	m.textInput.SetCursor(1)
}

func (m *Model) GetValue() (bson.D, error) {
	var query bson.D
	err := bson.UnmarshalExtJSON([]byte(m.textInput.Value()), false, &query)
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
