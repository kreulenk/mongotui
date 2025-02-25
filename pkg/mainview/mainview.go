// This package contains the main view for the TUI
// It has been separated from the tui package as the error modal needs to be templated over a valid tea.Model

package mainview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/dbcoltable"
	"github.com/kreulenk/mongotui/pkg/components/doclist"
	"github.com/kreulenk/mongotui/pkg/components/editor"
	"github.com/kreulenk/mongotui/pkg/components/jsonviewer"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
)

type Model struct {
	state *state.MainViewState

	dbColTable      *dbcoltable.Model
	docList         *doclist.Model
	singleDocViewer *jsonviewer.Model
	singleDocEditor editor.Editor

	engine *mongoengine.Engine
}

func New(engine *mongoengine.Engine) *Model {
	s := state.DefaultState()
	d := dbcoltable.New(engine, s) // This will be the first component to be focused on startup
	d.Focus()
	return &Model{
		state:           s,
		dbColTable:      d,
		docList:         doclist.New(engine, s),
		singleDocViewer: jsonviewer.New(engine, s),
		singleDocEditor: editor.New(engine, s),
		engine:          engine,
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		leftRightBorderWidth := 2
		topBottomBorderAndHelpHeight := 3
		m.dbColTable.SetWidth((msg.Width/3)-leftRightBorderWidth, msg.Width-leftRightBorderWidth)
		m.dbColTable.SetHeight(msg.Height - topBottomBorderAndHelpHeight)
		m.docList.SetWidth((msg.Width * 2 / 3) - leftRightBorderWidth)
		m.docList.SetHeight(msg.Height - topBottomBorderAndHelpHeight)

		m.singleDocViewer.SetWidth(msg.Width)
		m.singleDocViewer.SetHeight(msg.Height)
		return m, tea.ClearScreen // Necessary for resizes
	case modal.ExecCollDrop, modal.ExecDbDrop: // A deletion was confirmed via the modal component
		m.dbColTable, cmd = m.dbColTable.Update(msg)
		return m, cmd
	case modal.ExecDocDelete:
		m.docList, cmd = m.docList.Update(msg)
		return m, cmd
	}

	switch m.state.GetActiveComponent() {
	case state.DbColTable:
		m.dbColTable, cmd = m.dbColTable.Update(msg)
		cmds = append(cmds, cmd)
		if m.state.IsComponentActive(state.DocList) { // If the state switched, use a fresh docList
			m.docList.Focus()
		}
	case state.DocList:
		m.docList, cmd = m.docList.Update(msg)
		cmds = append(cmds, cmd)
		if m.state.IsComponentActive(state.DbColTable) {
			m.dbColTable.Focus()
		} else if m.state.IsComponentActive(state.SingleDocViewer) {
			if err := m.singleDocViewer.Focus(); err != nil {
				return m, modal.DisplayErrorModal(err)
			}
		} else if m.state.IsComponentActive(state.SingleDocEditor) {
			cmd = m.singleDocEditor.EditDoc()
			m.state.SetActiveComponent(state.DocList)
			cmds = append(cmds, cmd, tea.ClearScreen)
		} else if m.state.IsComponentActive(state.DocInsert) {
			cmd = m.singleDocEditor.InsertDoc()
			m.state.SetActiveComponent(state.DocList)
			cmds = append(cmds, cmd, tea.ClearScreen)
		}
	case state.SingleDocViewer:
		m.singleDocViewer, cmd = m.singleDocViewer.Update(msg)
		if m.state.IsComponentActive(state.DocList) {
			m.docList.Focus()
		}
		cmds = append(cmds, cmd)
	case state.SingleDocEditor: // This shouldn't happen
		panic("SingleDocEditor should only be selected after an update to DocList")
	default:
		panic("unhandled default case")
	}

	return m, tea.Batch(cmds...)
}

// Init is used to satisfy the tea.Model interface
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	if m.state.GetActiveComponent() == state.SingleDocViewer {
		return m.singleDocViewer.View()
	}
	tables := lipgloss.JoinHorizontal(lipgloss.Left, m.dbColTable.View(), m.docList.View())
	if m.state.GetActiveComponent() == state.DbColTable {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.dbColTable.HelpView())
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.docList.HelpView())
	}
}

func (m *Model) IsDbCollFilterOrSearchQueryFocused() bool {
	return m.dbColTable.IsFilterEnabled() || m.docList.IsSearchFocused()
}
