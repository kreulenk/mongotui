package errormodal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	windowWidth  int
	windowHeight int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	}

	return m, cmd
}

func (m *Model) View() string {
	foreStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("6")).
		Padding(0, 1).
		Width(40)

	boldStyle := lipgloss.NewStyle().Bold(true)
	title := boldStyle.Render("Error")
	content := "This is an example error message.\n\nPress <space> to close the window."

	return foreStyle.Render(title + "\n\n" + content)
}
