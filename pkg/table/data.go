package table

import (
	"fmt"
	"mtui/pkg/mongodata"
	"os"
	"slices"
)

// updateTableRows updates the rows in the table based on the dbData and the current cursorColumn and cursorRow
// Lots of opportunity for caching with how this function is handled/called, but I like the live data for now
func (m *Model) updateTableRows() {
	err := m.engine.SetCollectionsPerDb(m.selectedDb)
	if err != nil { // TODO handle errors better
		fmt.Printf("could not fetch collections: %v", err)
		os.Exit(1)
	}
	databaseNames := getSortedDatabasesByName(m.engine.Server.Databases)

	var newRows []Row
	if m.cursorColumn == databasesColumn {
		collectionsNames := getSortedCollectionsByName(m.engine.Server.Databases[m.SelectedCell()].Collections)
		for i, dbName := range databaseNames {
			if i < len(collectionsNames) {
				newRows = append(newRows, Row{dbName, collectionsNames[i]})
			} else {
				newRows = append(newRows, Row{dbName, ""})
			}
		}
	} else if m.cursorColumn == collectionsColumn {
		collectionsNames := getSortedCollectionsByName(m.engine.Server.Databases[m.selectedDb].Collections)
		if len(databaseNames) > len(collectionsNames) {
			for i, dbName := range databaseNames {
				if i < len(collectionsNames) {
					newRows = append(newRows, Row{dbName, collectionsNames[i]})
				} else {
					newRows = append(newRows, Row{dbName, ""})
				}
			}
		} else {
			for i := range collectionsNames {
				if i < len(databaseNames) {
					newRows = append(newRows, Row{databaseNames[i], collectionsNames[i]})
				} else {
					newRows = append(newRows, Row{"", collectionsNames[i]})
				}
			}
		}
	}
	m.rows = newRows
}

func getSortedDatabasesByName(databases map[string]mongodata.Database) []string {
	databaseNames := make([]string, 0, len(databases))
	for name := range databases {
		databaseNames = append(databaseNames, name)
	}
	slices.Sort(databaseNames)
	return databaseNames
}

// getSortedCollectionsByName returns a slice of collection names sorted alphabetically
func getSortedCollectionsByName(collections map[string]mongodata.Collection) []string {
	collectionNames := make([]string, 0, len(collections))
	for name := range collections {
		collectionNames = append(collectionNames, name)
	}
	slices.Sort(collectionNames)
	return collectionNames
}
