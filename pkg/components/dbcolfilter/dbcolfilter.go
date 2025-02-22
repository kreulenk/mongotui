package dbcolfilter

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	textInput textinput.Model
}

func New() *Model {
	ti := textinput.New()
	ti.Placeholder = "Filter"
	ti.CharLimit = 156
	ti.Focus()

	return &Model{
		textInput: ti,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

// SetWidth sets the width of the viewport of the dbcolfilter
func (m *Model) SetWidth(w int) {
	m.textInput.Width = w
}

func (m *Model) Focused() bool {
	return m.textInput.Focused()
}

func (m *Model) SetValue(s string) {
	m.textInput.SetValue(s)
}

func (m *Model) GetValue() string {
	return m.textInput.Value()
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyDown:
			m.textInput.Blur()
			return m, nil
		default:
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.textInput.View()
}
