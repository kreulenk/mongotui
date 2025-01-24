package modal

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/internal/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
)

type Model struct {
	state  *state.TuiState
	engine *mongoengine.Engine

	styles Styles
}

// New returns a modal component with the default styles applied
func New(state *state.TuiState, engine *mongoengine.Engine) *Model {
	return &Model{
		state:  state,
		engine: engine,
		styles: defaultStyles(),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	db := m.state.MainViewState.DbColTableState.GetSelectedDbName()
	col := m.state.MainViewState.DbColTableState.GetSelectedCollectionName()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.state.ModalState.GetError() != nil {
				m.state.ModalState.ClearError()
			} else if m.state.ModalState.IsDatabaseDeletionRequested() {
				if err := m.engine.DropDatabase(db); err != nil {
					m.state.ModalState.SetError(fmt.Errorf("failed to drop database '%s': %w", db, err))
				}
				m.state.ModalState.ResetDatabaseDeletionPromptRequest()
				m.state.MainViewState.DbColTableState.RequestRefreshAfterDatabaseDeletion()
			} else if m.state.ModalState.IsCollectionDeletionRequested() {
				if err := m.engine.DropCollection(db, col); err != nil {
					m.state.ModalState.SetError(fmt.Errorf("failed to drop collection '%s' within database '%s': %w", col, db, err))
				}
				m.state.ModalState.ResetCollectionDeletionPromptRequest()
				m.state.MainViewState.DbColTableState.RequestRefreshAfterCollectionDeletion()
			}
		default: // If it is a confirmation modal and enter was not selected, exit the modal with no actions performed
			if m.state.ModalState.IsDatabaseDeletionRequested() || m.state.ModalState.IsCollectionDeletionRequested() {
				m.state.ModalState.ResetDatabaseDeletionPromptRequest()
				m.state.ModalState.ResetCollectionDeletionPromptRequest()
			}
		}
	}
	return m, nil
}

func (m *Model) View() string {
	if m.state.ModalState.GetError() != nil {
		title := m.styles.ErrorHeader.Render("Error")
		return m.styles.Modal.Render(title + "\n\n" + m.state.ModalState.GetError().Error())
	} else if m.state.ModalState.IsDatabaseDeletionRequested() {
		title := m.styles.DeletionHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to delete the database %s?\nPress Enter to confirm.", title, m.state.MainViewState.DbColTableState.GetSelectedDbName())
		return m.styles.Modal.Render(msg)
	} else if m.state.ModalState.IsCollectionDeletionRequested() {
		title := m.styles.DeletionHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\nAre you sure you would like to delete the collection %s?\nPress Enter to confirm.", title, m.state.MainViewState.DbColTableState.GetSelectedCollectionName())
		return m.styles.Modal.Render(msg)
	}
	return ""
}
