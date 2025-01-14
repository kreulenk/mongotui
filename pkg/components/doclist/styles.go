package doclist

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Table       lipgloss.Style
	Doc         lipgloss.Style
	SelectedDoc lipgloss.Style
	DocText     lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		Table: lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("240")),
		SelectedDoc: lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57")),
		Doc: lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("240")),
		DocText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("71")),
	}
}
