// This package is the main entrypoint of the TUI for mongotui.
// It initializes the Data engine and loads on the various display elements.

package tui

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/internal/state"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mainview"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	overlay "github.com/rmhubbert/bubbletea-overlay"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

// baseModel implements tea.Model, and manages the browser UI.
type baseModel struct {
	state    *state.TuiState
	errModal tea.Model
	mainView tea.Model
	overlay  tea.Model
}

func Initialize(client *mongo.Client) {
	p := tea.NewProgram(initialModel(client))
	defer client.Disconnect(context.TODO())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(client *mongo.Client) tea.Model {
	state := state.DefaultState()
	engine := mongoengine.New(client, state)

	errModal := modal.New(state, engine)
	mainView := mainview.New(state, engine)
	view := overlay.New(
		errModal,
		mainView,
		overlay.Center,
		overlay.Center,
		0,
		0,
	)

	return &baseModel{
		state:    state,
		errModal: errModal,
		mainView: mainView,
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
	if m.isModalRequested() {
		mod, modCmd := m.errModal.Update(message)
		m.errModal = mod
		if m.state.MainViewState.DbColTableState.WasDatabaseDeletedViaModal() ||
			m.state.MainViewState.DbColTableState.WasCollectionDeletedViaModal() {
			m.mainView.(*mainview.Model).RefreshAfterDeletion() // type cast necessary due to bubbletea-overlay requiring tea.Model
		}
		cmds = append(cmds, modCmd)
	} else {
		mv, mvCmd := m.mainView.Update(message)
		m.mainView = mv
		cmds = append(cmds, mvCmd)
	}

	return m, tea.Batch(cmds...)
}

// View applies and styling and handles rendering the view. It partly implements the tea.Model
// interface.
func (m *baseModel) View() string {
	if m.isModalRequested() {
		return m.overlay.View()
	}
	return m.mainView.View()
}

func (m *baseModel) isModalRequested() bool {
	return m.state.ModalState.GetError() != nil ||
		m.state.ModalState.IsDatabaseDeletionRequested() ||
		m.state.ModalState.IsCollectionDeletionRequested()
}
