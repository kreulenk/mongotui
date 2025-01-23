// The state package intends to keep track of state fields that need to be referenced/set by several different
// components. State that is only between a parent component and a child component should not be tracked via this package

package state

type MainViewComponent int

const (
	DbColTable MainViewComponent = iota
	DocList
	SingleDocViewer
	SingleDocEditor
)

type DocListState struct {
	selectedDocIndex int
}

func (d *DocListState) GetSelectedDocumentIndex() int {
	return d.selectedDocIndex
}

// SetSelectedDocIndex is called in DocList component and then used in mongo engine component to query a specific doc
func (d *DocListState) SetSelectedDocIndex(i int) {
	d.selectedDocIndex = i
}

type DbColTableState struct {
	databaseName   string
	collectionName string
}

func (d *DbColTableState) IsCollectionSelected() bool {
	return d.collectionName != "" && d.databaseName != ""
}

func (d *DbColTableState) ClearCollectionSelection() {
	d.collectionName = ""
	d.databaseName = ""
}

func (d *DbColTableState) GetSelectedDbName() string {
	return d.databaseName
}

func (d *DbColTableState) GetSelectedCollectionName() string {
	return d.collectionName
}

func (d *DbColTableState) SetSelectedCollection(dbName, collectionName string) {
	d.databaseName = dbName
	d.collectionName = collectionName
}

type MainViewState struct {
	ActiveComponent MainViewComponent
	DocListState    *DocListState
	DbColTableState *DbColTableState
}

type TuiState struct {
	err           error // If err is set, the modal component will be displayed with the message
	displayModal  bool
	MainViewState MainViewState
}

func (t *TuiState) SetError(err error) {
	t.err = err
}

func (t *TuiState) GetError() error {
	return t.err
}

func (t *TuiState) ClearError() {
	t.err = nil
}

func DefaultState() *TuiState {
	return &TuiState{
		err:          nil,
		displayModal: false,
		MainViewState: MainViewState{
			ActiveComponent: DbColTable,
			DocListState: &DocListState{
				selectedDocIndex: 0,
			},
			DbColTableState: &DbColTableState{
				databaseName:   "",
				collectionName: "",
			},
		},
	}
}
