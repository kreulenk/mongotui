package table

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type dataEngine struct {
	client *mongo.Client
	dbData *mongoData
}

type baseModel struct {
	table  tableModel
	engine *dataEngine
	err    error // TODO handle how to display errors
}

func InitializeTui(client *mongo.Client) {
	p := tea.NewProgram(initialModel(client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(client *mongo.Client) baseModel {
	engine := &dataEngine{
		client: client,
		dbData: &mongoData{
			databases: make(map[string]mongoDatabase),
		},
	}

	err := engine.setDatabases()
	if err != nil {
		fmt.Printf("could not initialize data: %v", err)
		os.Exit(1)
	}

	t := New(
		engine,
		WithColumns(columns),
		WithFocused(true),
	)

	s := DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	t.updateTableRows()

	return baseModel{
		table:  t,
		engine: engine,
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
		for i := range m.table.Columns() {
			m.table.Columns()[i].Width = getColumnWidth(msg.Width, len(m.table.Columns()))
		}
		m.table.SetWidth(msg.Width - 6) // TODO look into a more intelligent way of getting this 6 value
		m.table.SetHeight(msg.Height - 4)
		return m, tea.ClearScreen // Necessary for resizes
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func getColumnWidth(windowWidth int, columns int) int {
	return (windowWidth - 8) / columns
}

func (m baseModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}
