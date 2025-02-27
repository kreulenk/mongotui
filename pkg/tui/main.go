// This package is the main entrypoint of the TUI for mongotui.
// It initializes the Documents engine and loads on the various display elements.

package tui

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mainview"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"github.com/muesli/termenv"
	overlay "github.com/rmhubbert/bubbletea-overlay"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

// baseModel implements tea.Model, and manages the browser UI.
type baseModel struct {
	msgModal tea.Model
	mainView tea.Model
	overlay  tea.Model

	engine *mongoengine.Engine
}

func Initialize(client *mongo.Client) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	p := tea.NewProgram(initialModel(client))
	defer client.Disconnect(context.Background())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel(client *mongo.Client) tea.Model {
	engine := mongoengine.New(client)

	msgModal := modal.New()
	mainView := mainview.New(engine)
	view := overlay.New(
		msgModal,
		mainView,
		overlay.Center,
		overlay.Center,
		0,
		0,
	)

	return &baseModel{
		msgModal: msgModal,
		mainView: mainView,
		overlay:  view,

		engine: engine,
	}
}

// Init initialises the baseModel on program load. It partly implements the tea.Model interface.
func (m *baseModel) Init() tea.Cmd {
	return nil
}

// Update handles event and manages internal state. It partly implements the tea.Model interface.
func (m *baseModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	// First see if we need to redirect to the msgModal
	// TODO find a simpler way of finding all modal messages
	case modal.ErrModalMsg, modal.DbCollInsertModalMsg, modal.CollDropModalMsg, modal.DbDropModalMsg, modal.DocDeleteModalMsg, modal.DocInsertModalMsg, modal.DocEditModalMsg:
		mod, modCmd := m.msgModal.Update(message)
		m.msgModal = mod
		return m, modCmd
	case modal.ExecDocEdit: // Edit does not require cursor updates so it can be executed from top tui component
		if err := m.engine.UpdateDocument(msg.OldDoc, msg.NewDoc); err != nil {
			return m, modal.DisplayErrorModal(err)
		}
		return m, m.engine.RerunLastCollectionQuery()
	case modal.ExecDocInsert: // Insert does not require cursor updates so it can be executed from top tui component
		if err := m.engine.InsertDocument(msg.Doc); err != nil {
			return m, modal.DisplayErrorModal(err)
		}
		return m, m.engine.RerunLastCollectionQuery()
	case modal.ExecDbCollInsert:
		if err := m.engine.InsertDatabaseAndCollection(msg.DatabaseName, msg.CollectionName); err != nil {
			return m, modal.DisplayErrorModal(err)
		}
		if err := m.engine.RefreshDbAndCollections(); err != nil {
			return m, modal.DisplayErrorModal(fmt.Errorf("error refreshing data after database and collection insertion: %w", err))
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !m.mainView.(*mainview.Model).IsDbCollFilterOrSearchQueryFocused() && !m.msgModal.(*modal.Model).IsTextInputFocused() {
				return m, tea.Quit
			}
		}
	}
	if m.msgModal.(*modal.Model).IsModalDisplaying() {
		mod, modCmd := m.msgModal.Update(message)
		m.msgModal = mod
		return m, modCmd
	}
	mv, mvCmd := m.mainView.Update(message)
	m.mainView = mv
	return m, mvCmd
}

// View applies and styling and handles rendering the view. It partly implements the tea.Model
// interface.
func (m *baseModel) View() string {
	if m.msgModal.(*modal.Model).IsModalDisplaying() {
		return m.overlay.View()
	}
	return m.mainView.View()
}
