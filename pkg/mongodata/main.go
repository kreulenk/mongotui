// mongodata handles the fetching and caching of data retrieved from the MongoDB database.

package mongodata

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"slices"
	"time"
)

const Timeout = 5 * time.Second

type Server struct {
	Databases map[string]Database
}

type Database struct {
	Collections map[string]Collection
}

type Collection struct {
	Data []bson.M
}

type selectedCollection struct {
	collectionName string
	databaseName   string
}

type Engine struct {
	Client             *mongo.Client
	Server             *Server
	selectedCollection selectedCollection
}

func (m *Engine) SetDatabases() error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	dbNames, err := m.Client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.Server.Databases = make(map[string]Database)
	for _, dbName := range dbNames {
		err := m.SetCollectionsPerDb(dbName)
		if err != nil {
			return err
		}
	}

	return nil

}

// SetCollectionsPerDb fetches the Collections for a given database along with the number of records in each collection
func (m *Engine) SetCollectionsPerDb(dbName string) error {
	db := m.Client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.Server.Databases[dbName] = Database{Collections: make(map[string]Collection)} // zero out the Data
	for _, collectionName := range collectionNames {
		m.Server.Databases[dbName].Collections[collectionName] = Collection{}
	}

	return nil
}

// SetSelectedCollection Allows parent components to set what data will be displayed within this component.
func (m *Engine) SetSelectedCollection(collectionName, databaseName string) {
	m.selectedCollection = selectedCollection{
		collectionName: collectionName,
		databaseName:   databaseName,
	}
}

func (m *Engine) IsCollectionSelected() bool {
	return m.selectedCollection.collectionName != "" && m.selectedCollection.databaseName != ""
}

// QueryCollection fetches all the data from a given collection in a given database
func (m *Engine) QueryCollection(query bson.D) error {
	if m.selectedCollection.databaseName == "" || m.selectedCollection.collectionName == "" {
		return fmt.Errorf("no collection selected") // This should never happen
	}

	db := m.Client.Database(m.selectedCollection.databaseName)
	coll := db.Collection(m.selectedCollection.collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	cur, err := coll.Find(ctx, query)
	if err != nil {
		return err
	}

	var data []bson.M
	if err = cur.All(ctx, &data); err != nil {
		return err
	}

	curDb, ok := m.Server.Databases[m.selectedCollection.databaseName]
	if ok {
		curDb.Collections[m.selectedCollection.collectionName] = Collection{Data: data}
	} else {
		return fmt.Errorf("no database is set") // should never happen
	}
	return nil
}

func (m *Engine) GetSelectedDocs() []bson.M {
	db, ok := m.Server.Databases[m.selectedCollection.databaseName]
	if ok {
		collection, ok := db.Collections[m.selectedCollection.collectionName]
		if ok {
			return collection.Data
		}
	}
	return []bson.M{}
}

func GetSortedDatabasesByName(databases map[string]Database) []string {
	databaseNames := make([]string, 0, len(databases))
	for name := range databases {
		databaseNames = append(databaseNames, name)
	}
	slices.Sort(databaseNames)
	return databaseNames
}

// GetSortedCollectionsByName returns a slice of collection names sorted alphabetically
func GetSortedCollectionsByName(collections map[string]Collection) []string {
	collectionNames := make([]string, 0, len(collections))
	for name := range collections {
		collectionNames = append(collectionNames, name)
	}
	slices.Sort(collectionNames)
	return collectionNames
}
