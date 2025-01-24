package state

type DocListState struct {
	selectedDocIndex int // TODO find a better way to share selected document between doclist, mongoengine, and state packages

}

func (d *DocListState) GetSelectedDocumentIndex() int {
	return d.selectedDocIndex
}

// SetSelectedDocIndex is called in DocList component and then used in mongo engine component to query a specific doc
func (d *DocListState) SetSelectedDocIndex(i int) {
	d.selectedDocIndex = i
}
