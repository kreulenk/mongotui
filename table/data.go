package table

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"slices"
	"time"
)

type mongoData struct {
	databases map[string]mongoDatabase
}

type mongoDatabase struct {
	collections map[string]mongoCollection
}

type mongoCollection struct {
	count int64
	data  []bson.M
}

func (m *dataEngine) setDatabases() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	dbNames, err := m.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.dbData.databases = make(map[string]mongoDatabase)
	for _, dbName := range dbNames {
		err := m.setCollectionsPerDb(dbName)
		if err != nil {
			return err
		}
	}

	return nil

}

// setCollectionsPerDb fetches the collections for a given database along with the number of records in each collection
func (m *dataEngine) setCollectionsPerDb(dbName string) error {
	if dbName == "" { // TODO see if better way of handling uninitialized data
		return nil
	}
	db := m.client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.dbData.databases[dbName] = mongoDatabase{collections: make(map[string]mongoCollection)} // zero out the data
	for _, collectionName := range collectionNames {
		coll := db.Collection(collectionName)
		c, err := coll.CountDocuments(context.Background(), bson.D{})
		if err != nil {
			return err
		}
		m.dbData.databases[dbName].collections[collectionName] = mongoCollection{count: c}
	}

	return nil
}

// updateTableRows updates the rows in the table based on the dbData and the current cursorColumn and cursorRow
// Lots of opportunity for caching with how this function is handled/called, but I like the live data for now
func (m *tableModel) updateTableRows() {
	err := m.engine.setCollectionsPerDb(m.SelectedCell())
	if err != nil { // TODO handle errors better
		fmt.Printf("could not fetch collections: %v", err)
		os.Exit(1)
	}
	databaseNames := getSortedDatabasesByName(m.engine.dbData.databases)
	collectionsNames := getSortedCollectionsByName(m.engine.dbData.databases[m.SelectedCell()].collections)

	var newRows []Row
	if m.cursorColumn == databasesColumn {
		for i, dbName := range databaseNames {
			if i < len(collectionsNames) {
				newRows = append(newRows, Row{dbName, collectionsNames[i]})
			} else {
				newRows = append(newRows, Row{dbName, ""})
			}
		}
	} else if m.cursorColumn == collectionsColumn {
		for i, collectionName := range collectionsNames {
			if i < len(databaseNames) {
				newRows = append(newRows, Row{databaseNames[i], collectionName})
			} else {
				newRows = append(newRows, Row{"", collectionName})
			}
		}
	}
	m.rows = newRows
}

func getSortedDatabasesByName(databases map[string]mongoDatabase) []string {
	databaseNames := make([]string, 0, len(databases))
	for name := range databases {
		databaseNames = append(databaseNames, name)
	}
	slices.Sort(databaseNames)
	return databaseNames
}

// getSortedCollectionsByName returns a slice of collection names sorted alphabetically
func getSortedCollectionsByName(collections map[string]mongoCollection) []string {
	collectionNames := make([]string, 0, len(collections))
	for name := range collections {
		collectionNames = append(collectionNames, name)
	}
	slices.Sort(collectionNames)
	return collectionNames
}
