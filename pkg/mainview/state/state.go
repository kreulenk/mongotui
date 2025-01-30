// The state package is used to keep track of which component is currently 'active' within the 'mainview' package
// of mongotui

package state

type ActiveComponent int

const (
	DbColTable ActiveComponent = iota
	DocList
	SingleDocViewer
	SingleDocEditor
)

func DefaultState() *MainViewState {
	return &MainViewState{
		activeComponent: DbColTable,
	}
}

type MainViewState struct {
	activeComponent ActiveComponent
}

func (m *MainViewState) SetActiveComponent(componentName ActiveComponent) {
	m.activeComponent = componentName
}

func (m *MainViewState) GetActiveComponent() ActiveComponent {
	return m.activeComponent
}
