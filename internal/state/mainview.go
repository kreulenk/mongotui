package state

type MainViewState struct {
	activeComponent MainViewComponent
	DocListState    *DocListState
	DbColTableState *DbColTableState
}

func (m *MainViewState) SetActiveComponent(componentName MainViewComponent) {
	m.activeComponent = componentName
}

func (m *MainViewState) GetActiveComponent() MainViewComponent {
	return m.activeComponent
}
