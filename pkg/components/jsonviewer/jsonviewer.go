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
	"github.com/kreulenk/mongotui/pkg/components/errormodal"
	"github.com/kreulenk/mongotui/pkg/mongodata"
)

type Model struct {
	Viewport viewport.Model
	Help     help.Model
	errModal *errormodal.Model

	engine *mongodata.Engine
	focus  bool
}

func New(engine *mongodata.Engine, errModal *errormodal.Model) *Model {
	viewPort := viewport.New(0, 0)

	return &Model{
		Viewport: viewPort,
		Help:     help.New(),
		engine:   engine,
		focus:    false,
		errModal: errModal,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Focus() {
	selectedDoc, err := m.engine.GetSelectedDocument()
	if err != nil {
		m.errModal.SetError(fmt.Errorf("could not get selected document: %v", err))
		return
	}

	buf := new(bytes.Buffer)
	if err := quick.Highlight(buf, selectedDoc, "json", "terminal256", "dracula"); err != nil {
		m.errModal.SetError(fmt.Errorf("could not highlight json: %v", err))
		return
	}

	renderedContent := lipgloss.NewStyle().
		Width(m.Viewport.Width).
		Height(m.Viewport.Height).
		Render(buf.String())

	m.Viewport.SetContent(renderedContent)
	m.focus = true
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) blur() {
	m.focus = false
	m.engine.ClearSelectedDocument()
}

func (m *Model) SetWidth(w int) {
	m.Viewport.Width = w - 1
}

func (m *Model) SetHeight(h int) {
	m.Viewport.Height = h - 1 // 1 line for help menu
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.blur()
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
