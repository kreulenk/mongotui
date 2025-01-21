package modal

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	promptMsg string
	err       error

	styles Styles
}

// New returns a modal component with the default styles applied
func New() *Model {
	return &Model{
		styles: defaultStyles(),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeySpace, tea.KeyEnter:
			if m.ShouldDisplay() {
				m.err = nil
				m.promptMsg = ""
			}
		default:
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if !m.ShouldDisplay() {
		return ""
	}
	if m.err != nil {
		title := m.styles.Header.Render("Error")
		return m.styles.Modal.Render(title + "\n\n" + m.err.Error())
	} else {
		return ""
	}
}

func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) SetPromptMessage(msg string) {
	m.promptMsg = msg
}

func (m *Model) ShouldDisplay() bool {
	return m.err != nil || m.promptMsg != ""
}
