package jsonviewer

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/errormodal"
	"github.com/kreulenk/mongotui/pkg/mongodata"
)

type Model struct {
	Viewport viewport.Model
	engine   *mongodata.Engine
	errModal *errormodal.Model
	focus    bool
}

func New(engine *mongodata.Engine, errModal *errormodal.Model) *Model {
	viewPort := viewport.New(0, 0)

	return &Model{
		Viewport: viewPort,
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
	if err := quick.Highlight(buf, selectedDoc, ".json", "terminal256", "dracula"); err != nil {
		m.errModal.SetError(fmt.Errorf("could not highlight json: %v", err))
		return
	}

	renderedContent := lipgloss.NewStyle().
		Width(m.Viewport.Width).
		Height(m.Viewport.Height).
		Render(selectedDoc)

	m.Viewport.SetContent(renderedContent)
	m.focus = true
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) blur() {
	m.focus = false
}

func (m *Model) SetWidth(w int) {
	m.Viewport.Width = w
}

func (m *Model) SetHeight(h int) {
	m.Viewport.Height = h
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.focus {
		var cmd tea.Cmd
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return m.Viewport.View()
}
