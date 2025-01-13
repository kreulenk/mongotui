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
	coltable *coltable.Model
	doclist  *doclist.Model

	errModal *errormodal.Model

	componentSelection componentSelection

	engine *mongodata.Engine
	err    error // TODO handle how to display errors
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
		fmt.Printf("could not initialize Data: %v", err)
		os.Exit(1)
	}

	t := coltable.New(engine, errModal)
	d := doclist.New(engine, errModal)

	return &Model{
		coltable:           t,
		doclist:            d,
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
		m.coltable.SetWidth((msg.Width / 3) - leftRightBorderWidth)
		m.coltable.SetHeight(msg.Height - topBottomBorderAndHelpHeight)

		m.doclist.SetWidth((msg.Width * 2 / 3) - leftRightBorderWidth)
		m.doclist.SetHeight(msg.Height - topBottomBorderAndHelpHeight)

		return m, tea.ClearScreen // Necessary for resizes
	}

	m.coltable, _ = m.coltable.Update(msg)
	m.doclist, _ = m.doclist.Update(msg)

	// If a collection is selected, pass off control to the doclist
	if m.componentSelection == dbColSelection && m.coltable.CollectionSelected() {
		m.componentSelection = docSummarySelection
		m.engine.SetSelectedCollection(m.coltable.SelectedCollection(), m.coltable.SelectedDatabase())
		m.doclist.Focus()
	}

	if m.componentSelection == docSummarySelection && m.doclist.Focused() == false {
		m.componentSelection = dbColSelection
		m.coltable.DeselectCollection()
	}

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	tables := lipgloss.JoinHorizontal(lipgloss.Left, m.coltable.View(), m.doclist.View())
	if m.coltable.CollectionSelected() {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.doclist.HelpView())
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.coltable.HelpView())
	}
}
