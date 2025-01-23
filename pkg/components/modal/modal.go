package modal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/internal/state"
)

type Model struct {
	promptMsg string
	state     *state.TuiState

	styles Styles
}

// New returns a modal component with the default styles applied
func New(state *state.TuiState) *Model {
	return &Model{
		state:  state,
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
			m.promptMsg = ""
			m.state.ClearError()
		default:
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if m.state.GetError() != nil {
		title := m.styles.Header.Render("Error")
		return m.styles.Modal.Render(title + "\n\n" + m.state.GetError().Error())
	} else {
		return ""
	}
}

func (m *Model) SetPromptMessage(msg string) {
	m.promptMsg = msg
}
