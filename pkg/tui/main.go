// This package is the main entrypoint of the TUI for mongotui.
// It initializes the Data engine and loads on the various display elements.

package tui

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/appview"
	"github.com/kreulenk/mongotui/pkg/components/modal"
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
	state    sessionState
	errModal tea.Model
	appView  tea.Model
	overlay  tea.Model
}

func Initialize(client *mongo.Client) {
	p := tea.NewProgram(initialModel(client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	if err := client.Disconnect(context.TODO()); err != nil {
		fmt.Printf("failed to properly disconnect from db. Closing anyways: %v\n", err)
	}
}

func initialModel(client *mongo.Client) tea.Model {
	errModal := modal.New()
	appView := appview.New(client, errModal)
	view := overlay.New(
		errModal,
		appView,
		overlay.Center,
		overlay.Center,
		0,
		0,
	)

	return &baseModel{
		state:    mainView,
		errModal: errModal,
		appView:  appView,
		overlay:  view,
	}
}

// Init initialises the baseModel on program load. It partly implements the tea.Model interface.
func (m *baseModel) Init() tea.Cmd {
	return nil
}

// Update handles event and manages internal state. It partly implements the tea.Model interface.
func (m *baseModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd
	if m.errModal.(*modal.Model).ShouldDisplay() { // TODO investigate if we NEED this typecast due to the bubble overlay library
		m.state = modalView
		fg, fgCmd := m.errModal.Update(message)
		m.errModal = fg
		cmds = append(cmds, fgCmd)
	} else {
		m.state = mainView
		bg, bgCmd := m.appView.Update(message)
		m.appView = bg
		cmds = append(cmds, bgCmd)

		// Check if the appView has an error
		if m.errModal.(*modal.Model).ShouldDisplay() {
			m.state = modalView
		}

	}

	return m, tea.Batch(cmds...)
}

// View applies and styling and handles rendering the view. It partly implements the tea.Model
// interface.
func (m *baseModel) View() string {
	if m.state == modalView {
		return m.overlay.View()
	}
	return m.appView.View()
}
