package table

import (
	"fmt"
	"mtui/pkg/mongodata"
	"os"
)

// updateTableData updates the data tracked in the model based on the current cursorDatabase, cursorCollection and cursorColumn position
// Lots of opportunity for caching with how this function is handled/called, but I like the live data for now
func (m *Model) updateTableData() {
	err := m.engine.SetCollectionsPerDb(m.databases[m.cursorDatabase])
	if err != nil { // TODO handle errors better
		fmt.Printf("could not fetch collections: %v", err)
		os.Exit(1)
	}

	m.collections = mongodata.GetSortedCollectionsByName(m.engine.Server.Databases[m.databases[m.cursorDatabase]].Collections)
}
