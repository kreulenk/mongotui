package modal

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
)

type Model struct {
	engine *mongoengine.Engine
	styles Styles

	errMsg       *ErrModalMsg // Sent in as a tea.Cmd from elsewhere in the program
	colDeleteMsg *ColDeleteModalMsg
}

// New returns a modal component with the default styles applied
func New(engine *mongoengine.Engine) *Model {
	return &Model{
		engine: engine,
		styles: defaultStyles(),

		errMsg: nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) IsModalDisplaying() bool {
	return m.errMsg != nil || m.colDeleteMsg != nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//db := m.state.MainViewState.DbColTableState.GetSelectedDbName()
	//col := m.state.MainViewState.DbColTableState.GetSelectedCollectionName()
	switch msg := msg.(type) {
	case ErrModalMsg:
		m.errMsg = &msg
	case ColDeleteModalMsg:
		m.colDeleteMsg = &msg
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
			}
		default: // If it is a confirmation modal and enter was not selected, exit the modal with no actions performed
			if m.colDeleteMsg != nil {
				m.colDeleteMsg = nil
			}
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if m.errMsg != nil {
		title := m.styles.ErrorHeader.Render("Error")
		return m.styles.Modal.Render(title + "\n\n" + m.errMsg.err.Error())
	} else if m.colDeleteMsg != nil {
		title := m.styles.DeletionHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\nAre you sure you would like to delete the collection %s?\nPress Enter to confirm.", title, m.colDeleteMsg.collectionName)
		return m.styles.Modal.Render(msg)
	}

	//if m.state.ModalState.GetError() != nil {
	//	title := m.styles.ErrorHeader.Render("Error")
	//	return m.styles.Modal.Render(title + "\n\n" + m.state.ModalState.GetError().Error())
	//} else if m.state.ModalState.IsDatabaseDeletionRequested() {
	//	title := m.styles.DeletionHeader.Render("Confirm")
	//	msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to delete the database %s?\nPress Enter to confirm.", title, m.state.MainViewState.DbColTableState.GetSelectedDbName())
	//	return m.styles.Modal.Render(msg)
	//} else if m.state.ModalState.IsCollectionDeletionRequested() {
	//	title := m.styles.DeletionHeader.Render("Confirm")
	//	msg := fmt.Sprintf("%s\n\nAre you sure you would like to delete the collection %s?\nPress Enter to confirm.", title, m.state.MainViewState.DbColTableState.GetSelectedCollectionName())
	//	return m.styles.Modal.Render(msg)
	//}
	return ""
}
