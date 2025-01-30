// This package contains the main view for the TUI
// It has been separated from the tui package as the error modal needs to be templated over a valid tea.Model

package mainview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/internal/state"
	"github.com/kreulenk/mongotui/pkg/components/dbcoltable"
	"github.com/kreulenk/mongotui/pkg/components/doclist"
	"github.com/kreulenk/mongotui/pkg/components/editor"
	"github.com/kreulenk/mongotui/pkg/components/jsonviewer"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
)

type Model struct {
	state *state.TuiState

	dbColTable      *dbcoltable.Model
	docList         *doclist.Model
	singleDocViewer *jsonviewer.Model
	singleDocEditor editor.Editor

	engine *mongoengine.Engine
}

func New(state *state.TuiState, engine *mongoengine.Engine) *Model {
	d := dbcoltable.New(engine, state) // This will be the first component to be focused on startup
	d.Focus()
	return &Model{
		state:           state,
		dbColTable:      d,
		docList:         doclist.New(engine, state),
		singleDocViewer: jsonviewer.New(engine, state),
		singleDocEditor: editor.New(engine, state),
		engine:          engine,
	}
}

// RefreshAfterDeletion is used after a deletion has been confirmed and successfully
// performed by the modal component so that the component with the deleted data can be
// properly updated
func (m *Model) RefreshAfterDeletion() error {
	switch m.state.MainViewState.GetActiveComponent() {
	case state.DbColTable:
		return m.dbColTable.RefreshAfterDeletion()
	default:
		panic("unhandled default case")
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		leftRightBorderWidth := 2
		topBottomBorderAndHelpHeight := 3
		m.dbColTable.SetWidth((msg.Width / 3) - leftRightBorderWidth)
		m.dbColTable.SetHeight(msg.Height - topBottomBorderAndHelpHeight)
		m.docList.SetWidth((msg.Width * 2 / 3) - leftRightBorderWidth)
		m.docList.SetHeight(msg.Height - topBottomBorderAndHelpHeight)

		m.singleDocViewer.SetWidth(msg.Width)
		m.singleDocViewer.SetHeight(msg.Height)

		return m, tea.ClearScreen // Necessary for resizes
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch m.state.MainViewState.GetActiveComponent() {
	case state.DbColTable:
		m.dbColTable, cmd = m.dbColTable.Update(msg)
		if m.state.MainViewState.GetActiveComponent() == state.DocList { // If the state switched, use a fresh docList
			m.docList.ResetSearchBar()
			err := m.docList.RefreshDocs()
			if err != nil {
				return m, modal.DisplayErrorModal(err)
			}
			m.docList.Focus()
		}
		cmds = append(cmds, cmd)
	case state.DocList:
		m.docList, cmd = m.docList.Update(msg)
		if m.state.MainViewState.GetActiveComponent() == state.DbColTable {
			m.dbColTable.Focus()
		} else if m.state.MainViewState.GetActiveComponent() == state.SingleDocViewer {
			if err := m.singleDocViewer.Focus(); err != nil {
				return m, modal.DisplayErrorModal(err)
			}
		} else if m.state.MainViewState.GetActiveComponent() == state.SingleDocEditor {
			if err := m.singleDocEditor.OpenFileInEditor(); err != nil {
				return m, modal.DisplayErrorModal(err)
			}
			if err := m.docList.RefreshDocs(); err != nil {
				return m, modal.DisplayErrorModal(err)
			}
		}
		cmds = append(cmds, cmd)
	case state.SingleDocViewer:
		m.singleDocViewer, cmd = m.singleDocViewer.Update(msg)
		if m.state.MainViewState.GetActiveComponent() == state.DocList {
			if err := m.docList.RefreshDocs(); err != nil {
				return m, modal.DisplayErrorModal(err)
			}
			m.docList.Focus()
		}
		cmds = append(cmds, cmd)
	case state.SingleDocEditor:
		panic("SingleDocEditor should only be selected after an update to DocList")
	default:
		panic("unhandled default case")
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	if m.state.MainViewState.GetActiveComponent() == state.SingleDocViewer {
		return m.singleDocViewer.View()
	}
	tables := lipgloss.JoinHorizontal(lipgloss.Left, m.dbColTable.View(), m.docList.View())
	if m.state.MainViewState.GetActiveComponent() == state.DbColTable {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.dbColTable.HelpView())
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.docList.HelpView())

	}
}
