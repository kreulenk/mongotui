// The state package intends to keep track of state fields that need to be referenced/set by several different
// components. State that is only between a parent component and a child component should not be tracked via this package

// Any component can query the state of any other component

package state

type MainViewComponent int

const (
	DbColTable MainViewComponent = iota
	DocList
	SingleDocViewer
	SingleDocEditor
)

type TuiState struct {
	MainViewState *MainViewState
}

func DefaultState() *TuiState {
	return &TuiState{
		MainViewState: &MainViewState{
			activeComponent: DbColTable,
			DbColTableState: &DbColTableState{
				databaseName:   "",
				collectionName: "",
			},
		},
	}
}
