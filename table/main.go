package table

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func InitializeTui(client *mongo.Client) {
	p := tea.NewProgram(initialModel(client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(client *mongo.Client) model {
	listCtx, listCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer listCancel()
	dbNames, err := client.ListDatabaseNames(listCtx, bson.D{})
	if err != nil {
		fmt.Printf("could not fetch databases: %v", err)
		os.Exit(1)
	}

	w, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		w = 80
	}
	colWidth := (w - 6) / 3
	columns := []table.Column{
		{Title: "Databases", Width: colWidth},
		{Title: "Collections", Width: colWidth},
		{Title: "Records", Width: colWidth},
	}

	var rows []table.Row
	for _, dbName := range dbNames {
		db := client.Database(dbName)
		collNames, err := db.ListCollectionNames(context.Background(), bson.D{})
		if err != nil {
			fmt.Printf("could not fetch collections: %v", err)
			os.Exit(1)
		}
		for _, collName := range collNames {
			coll := db.Collection(collName)
			count, err := coll.CountDocuments(context.Background(), bson.D{})
			if err != nil {
				fmt.Printf("could not fetch count: %v", err)
				os.Exit(1)
			}
			rows = append(rows, table.Row{dbName, collName, fmt.Sprintf("%d", count)})
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(h-6), // TODO look into a more intelligent way of getting this 6 value
		table.WithWidth(w-6),
	)

	s := table.DefaultStyles()
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

	return model{
		table: t,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.table.Columns()[i].Width = (msg.Width - 6) / 3
		}
		m.table.SetWidth(msg.Width - 6) // TODO look into a more intelligent way of getting this 6 value
		m.table.SetHeight(msg.Height - 6)
		return m, tea.ClearScreen // Necessary for resizes
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}
