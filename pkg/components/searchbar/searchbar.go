package searchbar

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"mtui/pkg/mongodata"
)

type Model struct {
	textInput textinput.Model
	engine    *mongodata.Engine
}

func New(engine *mongodata.Engine) Model {
	ti := textinput.New()
	ti.Placeholder = "Query"
	ti.CharLimit = 156
	ti.Width = 20

	return Model{
		textInput: ti,
		engine:    engine,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// SetWidth sets the width of the viewport of the coltable.
func (m *Model) SetWidth(w int) {
	m.textInput.Width = w
}

func (m *Model) Focus() {
	m.textInput.Focus()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.textInput.Blur()
			return m, nil
		}

	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.textInput.View()
}
