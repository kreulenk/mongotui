package dbcoltable

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Table    lipgloss.Style
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		Table: lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("57")),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false),
		Header: lipgloss.NewStyle().
			Inline(true).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true),
		Cell: lipgloss.NewStyle().
			Inline(true),
	}
}
