package errormodal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	err error
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			if m.ErrorPresent() {
				m.err = nil
			}
		}
	}
	return m, nil
}

func (m *Model) View() string {
	foreStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("6")).
		Padding(0, 1).
		Width(40)

	boldStyle := lipgloss.NewStyle().Bold(true)
	title := boldStyle.Render("Error")

	return foreStyle.Render(title + "\n\n" + m.err.Error())
}

func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) ErrorPresent() bool {
	return m.err != nil
}

func (m *Model) ClearError() {
	m.err = nil
}
