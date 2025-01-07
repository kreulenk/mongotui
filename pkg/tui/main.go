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

// TODO maybe make a style package so that all componenets can share the same styles
//var baseStyle = lipgloss.NewStyle().
//	BorderStyle(lipgloss.NormalBorder()).
//	BorderForeground(lipgloss.Color("240"))

type baseModel struct {
	table   table.Model
	doclist doclist.Model
	engine  *mongodata.Engine
	err     error // TODO handle how to display errors
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

	t := table.New(
		engine,
		table.WithFocused(true),
	)

	// TODO move these styles to the table package
	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	ts.Selected = ts.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(ts)

	d := doclist.New(engine)

	return baseModel{
		table:   t,
		doclist: d,
		engine:  engine,
	}
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc": // TODO use the focus/blur for when we are opening any modals
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		splitWidth := msg.Width / 2
		m.table.SetWidth(splitWidth - 6) // TODO look into a more intelligent way of getting this 6 value
		m.table.SetHeight(msg.Height - 4)

		m.doclist.SetWidth(splitWidth - 6)
		m.doclist.SetHeight(msg.Height - 4)

		return m, tea.ClearScreen // Necessary for resizes
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m baseModel) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.JoinVertical(lipgloss.Top, m.table.View(), m.table.HelpView()), m.doclist.View())
}
