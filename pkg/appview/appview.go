// This package contains the main view for the TUI
// It has been separated from the tui package as the error modal needs to be templated over a valid tea.Model

package appview

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/dbcoltable"
	"github.com/kreulenk/mongotui/pkg/components/doclist"
	"github.com/kreulenk/mongotui/pkg/components/editor"
	"github.com/kreulenk/mongotui/pkg/components/jsonviewer"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mongodata"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

type Model struct {
	dbColTable      *dbcoltable.Model
	docList         *doclist.Model
	singleDocViewer *jsonviewer.Model
	singleDocEditor editor.Editor
	msgModal        *modal.Model

	engine *mongodata.Engine
}

func New(client *mongo.Client, errModal *modal.Model) *Model {
	engine := mongodata.New(client)
	err := engine.SetDatabases()
	if err != nil {
		fmt.Printf("could not initialize data: %v", err)
		os.Exit(1)
	}

	return &Model{
		dbColTable:      dbcoltable.New(engine, errModal),
		docList:         doclist.New(engine, errModal),
		singleDocViewer: jsonviewer.New(engine, errModal),
		singleDocEditor: editor.New(engine),
		msgModal:        errModal,
		engine:          engine,
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

	m.dbColTable, _ = m.dbColTable.Update(msg)
	m.docList, _ = m.docList.Update(msg)
	if m.engine.IsDocumentSelected() && m.docList.GetDocAction() == doclist.EditDoc {
		if err := m.singleDocEditor.OpenFileInEditor(); err != nil {
			m.msgModal.SetError(err)
		}
	}

	if m.engine.IsDocumentSelected() && m.docList.GetDocAction() == doclist.ViewDoc { // Handle viewing of a single document
		if m.singleDocViewer.Focused() == false {
			m.singleDocViewer.Focus()
		}
		m.singleDocViewer, _ = m.singleDocViewer.Update(msg)
		if m.singleDocViewer.Focused() == false {
			m.docList.Focus(false)
		}
		return m, nil
	}

	// If a collection is selected, pass off control to the docList
	if !m.docList.Focused() && m.engine.IsCollectionSelected() {
		m.docList.Focus(true)
	}

	if !m.dbColTable.Focused() && !m.engine.IsCollectionSelected() {
		m.dbColTable.Focus()
	}

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	if m.singleDocViewer.Focused() {
		return m.singleDocViewer.View()
	}
	tables := lipgloss.JoinHorizontal(lipgloss.Left, m.dbColTable.View(), m.docList.View())
	if m.dbColTable.Focused() {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.dbColTable.HelpView())
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.docList.HelpView())

	}
}
