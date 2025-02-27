package modal

import (
	tea "github.com/charmbracelet/bubbletea"
)

type confirmationButtonCursor int

const (
	yesButtonCursor confirmationButtonCursor = iota
	noButtonCursor
)

type Model struct {
	styles Styles

	errMsg       *ErrModalMsg // Sent in as a tea.Cmd from elsewhere in the program
	collDropMsg  *CollDropModalMsg
	dbDropMsg    *DbDropModalMsg
	docDeleteMsg *DocDeleteModalMsg
	docInsertMsg *DocInsertModalMsg
	docEditMsg   *DocEditModalMsg

	confirmationCursor confirmationButtonCursor
}

// New returns a modal component with the default styles applied
func New() *Model {
	return &Model{
		styles: defaultStyles(),

		errMsg:       nil,
		collDropMsg:  nil,
		dbDropMsg:    nil,
		docDeleteMsg: nil,
		docInsertMsg: nil,
		docEditMsg:   nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) IsModalDisplaying() bool {
	return m.errMsg != nil ||
		m.collDropMsg != nil ||
		m.dbDropMsg != nil ||
		m.docDeleteMsg != nil ||
		m.docInsertMsg != nil ||
		m.docEditMsg != nil
}
