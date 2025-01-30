// The state package intends to keep track of state fields that need to be referenced/set by several different
// components. State that is only between a parent component and a child component should not be tracked via this package

// Any component can query the state of any other component

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
