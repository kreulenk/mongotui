package modal

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	styles Styles

	errMsg       *ErrModalMsg // Sent in as a tea.Cmd from elsewhere in the program
	colDeleteMsg *ColDeleteModalMsg
	dbDeleteMsg  *DbDeleteModalMsg
	docDeleteMsg *DocDeleteModalMsg
	docInsertMsg *DocInsertModalMsg
	docEditMsg   *DocEditModalMsg
}

// New returns a modal component with the default styles applied
func New() *Model {
	return &Model{
		styles: defaultStyles(),

		errMsg:       nil,
		colDeleteMsg: nil,
		dbDeleteMsg:  nil,
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
		m.colDeleteMsg != nil ||
		m.dbDeleteMsg != nil ||
		m.docDeleteMsg != nil ||
		m.docInsertMsg != nil ||
		m.docEditMsg != nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ErrModalMsg:
		m.errMsg = &msg
	case ColDeleteModalMsg:
		m.colDeleteMsg = &msg
	case DbDeleteModalMsg:
		m.dbDeleteMsg = &msg
	case DocDeleteModalMsg:
		m.docDeleteMsg = &msg
	case DocInsertModalMsg:
		m.docInsertMsg = &msg
	case DocEditModalMsg:
		m.docEditMsg = &msg
	case tea.KeyMsg:
		if !m.IsModalDisplaying() {
			return m, nil
		}
		switch msg.Type {
		case tea.KeyEnter:
			if m.errMsg != nil {
				m.errMsg = nil
			} else if m.colDeleteMsg != nil {
				execCmd := execCollectionDelete(m.colDeleteMsg.dbName, m.colDeleteMsg.collectionName)
				m.colDeleteMsg = nil
				return m, execCmd
			} else if m.dbDeleteMsg != nil {
				execCmd := execDatabaseDelete(m.dbDeleteMsg.dbName)
				m.dbDeleteMsg = nil
				return m, execCmd
			} else if m.docDeleteMsg != nil {
				execCmd := execDocDelete(m.docDeleteMsg.doc)
				m.docDeleteMsg = nil
				return m, execCmd
			} else if m.docInsertMsg != nil {
				execCmd := execDocInsert(m.docInsertMsg.doc)
				m.docInsertMsg = nil
				return m, execCmd
			} else if m.docEditMsg != nil {
				execCmd := execDocEdit(m.docEditMsg.oldDoc, m.docEditMsg.newDoc)
				m.docEditMsg = nil
				return m, execCmd
			}
		default: // If it is a confirmation modal and enter was not selected, exit the modal with no actions performed
			m.errMsg = nil
			m.colDeleteMsg = nil
			m.dbDeleteMsg = nil
			m.docDeleteMsg = nil
			m.docInsertMsg = nil
			m.docEditMsg = nil
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if m.errMsg != nil {
		title := m.styles.ErrorHeader.Render("Error")
		return m.styles.Modal.Render(title + "\n\n" + m.errMsg.Err.Error())
	} else if m.colDeleteMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\nAre you sure you would like to delete the collection %s?\nPress Enter to confirm.", title, m.colDeleteMsg.collectionName)
		return m.styles.Modal.Render(msg)
	} else if m.dbDeleteMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to delete the database %s?\nPress Enter to confirm.", title, m.dbDeleteMsg.dbName)
		return m.styles.Modal.Render(msg)
	} else if m.docDeleteMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to delete the selected document?\nPress Enter to confirm.", title)
		return m.styles.Modal.Render(msg)
	} else if m.docInsertMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to insert the document you just created?\nPress Enter to confirm.", title)
		return m.styles.Modal.Render(msg)
	} else if m.docEditMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to make the edits you just made?\nPress Enter to confirm.", title)
		return m.styles.Modal.Render(msg)
	}
	return ""
}
