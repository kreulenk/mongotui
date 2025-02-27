package modal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type confirmationButtonCursor int

const (
	yesButtonCursor confirmationButtonCursor = iota
	noButtonCursor
)

type dbCollInput int

const (
	databaseInputFocused dbCollInput = iota
	collectionInputFocused
)

type Model struct {
	styles Styles

	errMsg *ErrModalMsg // Sent in as a tea.Cmd from elsewhere in the program

	dbCollInsertMsg *DbCollInsertModalMsg

	collDropMsg  *CollDropModalMsg
	dbDropMsg    *DbDropModalMsg
	docDeleteMsg *DocDeleteModalMsg
	docInsertMsg *DocInsertModalMsg
	docEditMsg   *DocEditModalMsg

	confirmationCursor confirmationButtonCursor

	dbInsertInput      textinput.Model
	collInsertInput    textinput.Model
	focusedDbCollInput dbCollInput
}

// New returns a modal component with the default styles applied
func New() *Model {
	dbTi := textinput.New()
	dbTi.Placeholder = "Database"
	dbTi.Focus()
	collTi := textinput.New()
	collTi.Placeholder = "Collection"
	collTi.Focus()

	return &Model{
		styles: defaultStyles(),

		errMsg:          nil,
		dbCollInsertMsg: nil,
		collDropMsg:     nil,
		dbDropMsg:       nil,
		docDeleteMsg:    nil,
		docInsertMsg:    nil,
		docEditMsg:      nil,

		confirmationCursor: yesButtonCursor,

		dbInsertInput:      dbTi,
		collInsertInput:    collTi,
		focusedDbCollInput: databaseInputFocused,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) IsModalDisplaying() bool {
	return m.errMsg != nil ||
		m.dbCollInsertMsg != nil ||
		m.collDropMsg != nil ||
		m.dbDropMsg != nil ||
		m.docDeleteMsg != nil ||
		m.docInsertMsg != nil ||
		m.docEditMsg != nil
}

// IsTextInputFocused is used to determine if the 'q' key should quit the app or be routed
// onto the modal component and then the dbInsertInput
func (m *Model) IsTextInputFocused() bool {
	return m.dbCollInsertMsg != nil
}
