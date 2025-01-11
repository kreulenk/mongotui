// This package is the main entrypoint of the TUI for mongotui.
// It initializes the Data engine and loads on the various display elements.

package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/mongo"
	"mtui/pkg/doclist"
	"mtui/pkg/mongodata"
	"mtui/pkg/table"
	"os"
)

type componentSelection int

const (
	dbColSelection componentSelection = iota
	docSummarySelection
)

type baseModel struct {
	table   table.Model
	doclist doclist.Model

	componentSelection componentSelection

	engine *mongodata.Engine
	err    error // TODO handle how to display errors
}

func Initialize(client *mongo.Client) {
	p := tea.NewProgram(initialModel(client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(client *mongo.Client) baseModel {
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

	t := table.New(engine)
	d := doclist.New(engine)

	return baseModel{
		table:              t,
		doclist:            d,
		componentSelection: dbColSelection,
		engine:             engine,
	}
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		leftRightBorderWidth := 4
		topBottomBorderAndHelpHeight := 3
		m.table.SetWidth((msg.Width / 3) - leftRightBorderWidth)
		m.table.SetHeight(msg.Height - topBottomBorderAndHelpHeight)

		m.doclist.SetWidth((msg.Width * 2 / 3) - leftRightBorderWidth)
		m.doclist.SetHeight(msg.Height - topBottomBorderAndHelpHeight)

		return m, tea.ClearScreen // Necessary for resizes
	}

	m.table, _ = m.table.Update(msg)
	m.doclist, _ = m.doclist.Update(msg)

	// If a collection is selected, pass off control to the doclist
	if m.componentSelection == dbColSelection && m.table.CollectionSelected() {
		m.componentSelection = docSummarySelection
		m.doclist.SetSelectedCollection(m.table.SelectedCollection(), m.table.SelectedDatabase())
		m.doclist.Focus()
	}

	if m.componentSelection == docSummarySelection && m.doclist.Focused() == false {
		m.componentSelection = dbColSelection
		m.table.DeselectCollection()
	}

	return m, nil
}

func (m baseModel) View() string {
	tables := lipgloss.JoinHorizontal(lipgloss.Left, m.table.View(), m.doclist.View())
	if m.table.CollectionSelected() {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.doclist.HelpView())
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, tables, m.table.HelpView())
	}
}
