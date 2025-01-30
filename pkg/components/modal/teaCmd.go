package modal

import tea "github.com/charmbracelet/bubbletea"

type ErrMsg struct {
	err error
}

func DisplayErrorModal(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrMsg{err: err}
	}
}
