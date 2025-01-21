package modal

import "github.com/charmbracelet/lipgloss"

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Modal  lipgloss.Style
	Header lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("6")).
			Padding(0, 1).
			Width(40),
		Header: lipgloss.NewStyle().Bold(true),
	}
}
