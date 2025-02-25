package doclist

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/querysearch"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
)

type Model struct {
	state *state.MainViewState
	Help  help.Model

	searchBar *querysearch.Model
	styles    Styles
	focused   bool

	cursor   int
	viewport viewport.Model

	engine *mongoengine.Engine
}

// New creates a new baseModel for the dbcoltable widget.
func New(engine *mongoengine.Engine, state *state.MainViewState) *Model {
	m := Model{
		state:     state,
		Help:      help.New(),
		searchBar: querysearch.New(),

		viewport: viewport.New(0, 20),
		styles:   defaultStyles(),
		focused:  false,

		engine: engine,
	}
	return &m
}

func (m *Model) Focus() {
	m.focused = true
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("57"))
}

func (m *Model) blur() {
	m.focused = false
	m.cursor = 0
	m.styles.Table = m.styles.Table.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
}

func (m *Model) IsSearchFocused() bool {
	return m.searchBar.Focused()
}
