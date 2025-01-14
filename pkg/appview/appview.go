// This package contains the main view for the TUI
// It has been separated from the tui package as the error modal needs to be templated over a valid tea.Model

package appview

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/coltable"
	"github.com/kreulenk/mongotui/pkg/components/doclist"
	"github.com/kreulenk/mongotui/pkg/components/errormodal"
	"github.com/kreulenk/mongotui/pkg/components/jsonviewer"
	"github.com/kreulenk/mongotui/pkg/mongodata"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

type componentSelection int

const (
	dbColSelection componentSelection = iota
	docSummarySelection
)

type Model struct {
	colTable        *coltable.Model
	docList         *doclist.Model
	singleDocViewer *jsonviewer.Model

	errModal *errormodal.Model

	componentSelection componentSelection

	engine *mongodata.Engine
}

func New(client *mongo.Client, errModal *errormodal.Model) *Model {
	engine := &mongodata.Engine{
		Client: client,
		Server: &mongodata.Server{
			Databases: make(map[string]mongodata.Database),
		},
	}

	err := engine.SetDatabases()
	if err != nil {
		fmt.Printf("could not initialize data: %v", err)
		os.Exit(1)
	}

	t := coltable.New(engine, errModal)
	d := doclist.New(engine, errModal)
	sdv := jsonviewer.New(engine, errModal)

	return &Model{
		colTable:           t,
		docList:            d,
		singleDocViewer:    sdv,
		errModal:           errModal,
		componentSelection: dbColSelection,
		engine:             engine,
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		leftRightBorderWidth := 4
		topBottomBorderAndHelpHeight := 3
		m.colTable.SetWidth((msg.Width / 3) - leftRightBorderWidth)
		m.colTable.SetHeight(msg.Height - topBottomBorderAndHelpHeight)
		m.docList.SetWidth((msg.Width * 2 / 3) - leftRightBorderWidth)
		m.docList.SetHeight(msg.Height - topBottomBorderAndHelpHeight)

		m.singleDocViewer.SetWidth(msg.Width)
		m.singleDocViewer.SetHeight(msg.Height)

		return m, tea.ClearScreen // Necessary for resizes
	}

	m.colTable, _ = m.colTable.Update(msg)
	m.docList, _ = m.docList.Update(msg)
	m.singleDocViewer, _ = m.singleDocViewer.Update(msg)

	if m.engine.IsDocumentSelected() {
		m.singleDocViewer.Focus()
		return m, nil
	}

	// If a collection is selected, pass off control to the docList
	if m.componentSelection == dbColSelection && m.colTable.CollectionSelected() {
		m.componentSelection = docSummarySelection
		m.engine.SetSelectedCollection(m.colTable.SelectedCollection(), m.colTable.SelectedDatabase())
		m.docList.Focus()
	}

	if m.componentSelection == docSummarySelection && m.docList.Focused() == false {
		m.componentSelection = dbColSelection
		m.colTable.DeselectCollection()
	}

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string { // TODO standardize how component focusing and blurring is handled
	if m.engine.IsDocumentSelected() {
		return m.singleDocViewer.View()
	}
	tables := lipgloss.JoinHorizontal(lipgloss.Left, m.colTable.View(), m.docList.View())
	if m.colTable.CollectionSelected() {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.docList.HelpView())
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.colTable.HelpView())
	}
}
