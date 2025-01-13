// This package is the main entrypoint of the TUI for mongotui.
// It initializes the Data engine and loads on the various display elements.

package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/appview"
	"github.com/kreulenk/mongotui/pkg/components/errormodal"
	overlay "github.com/rmhubbert/bubbletea-overlay"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

type sessionState int

const (
	mainView sessionState = iota
	modalView
)

// baseModel implements tea.Model, and manages the browser UI.
type baseModel struct {
	state        sessionState
	windowWidth  int
	windowHeight int
	foreground   tea.Model
	background   tea.Model
	overlay      tea.Model
}

func Initialize(client *mongo.Client) {
	p := tea.NewProgram(initialModel(client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(client *mongo.Client) tea.Model {
	appView := appview.New(client)
	modal := &errormodal.Model{}
	view := overlay.New(
		modal,
		appView,
		overlay.Center,
		overlay.Center,
		0,
		0,
	)

	return &baseModel{
		state:        mainView,
		windowWidth:  0,
		windowHeight: 0,
		foreground:   modal,
		background:   appView,
		overlay:      view,
	}
}

// Init initialises the baseModel on program load. It partly implements the tea.Model interface.
func (m *baseModel) Init() tea.Cmd {
	return nil
}

// Update handles event and manages internal state. It partly implements the tea.Model interface.
func (m *baseModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit

		case " ":
			if m.state == mainView {
				m.state = modalView
			} else {
				m.state = mainView
			}
			return m, nil
		}
	}

	fg, fgCmd := m.foreground.Update(message)
	m.foreground = fg

	bg, bgCmd := m.background.Update(message)
	m.background = bg

	cmds := []tea.Cmd{}
	cmds = append(cmds, fgCmd, bgCmd)

	return m, tea.Batch(cmds...)
}

// View applies and styling and handles rendering the view. It partly implements the tea.Model
// interface.
func (m *baseModel) View() string {
	if m.state == modalView {
		return m.overlay.View()
	}
	return m.background.View()
}
