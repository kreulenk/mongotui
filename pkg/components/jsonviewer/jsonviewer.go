package jsonviewer

import (
	"bytes"
	"fmt"

	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
)

type Model struct {
	state    *state.MainViewState
	Viewport viewport.Model
	Help     help.Model

	engine *mongoengine.Engine
}

func New(engine *mongoengine.Engine, state *state.MainViewState) *Model {
	viewPort := viewport.New(0, 0)

	return &Model{
		state:    state,
		Viewport: viewPort,
		Help:     help.New(),
		engine:   engine,
	}
}

func (m *Model) Focus() error {
	m.Viewport.GotoTop()
	selectedDoc, err := m.engine.GetSelectedDocumentMarshalled()
	if err != nil {
		return fmt.Errorf("could not fetch selected document: %v", err)
	}

	buf := new(bytes.Buffer)
	if err := quick.Highlight(buf, string(selectedDoc), "json", "terminal256", "dracula"); err != nil {
		return fmt.Errorf("could not highlight json: %v", err)
	}

	renderedContent := lipgloss.NewStyle().
		Width(m.Viewport.Width).
		Height(m.Viewport.Height).
		Render(buf.String())

	m.Viewport.SetContent(renderedContent)
	return nil
}

func (m *Model) SetWidth(w int) {
	m.Viewport.Width = w - 1
}

func (m *Model) SetHeight(h int) {
	m.Viewport.Height = h - 1 // 1 line for help menu
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.state.SetActiveComponent(state.DocList)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, m.Viewport.View(), m.Help.View(keys))
}
