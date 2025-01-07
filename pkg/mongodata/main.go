// mongodata handles the fetching and caching of data retrieved from the MongoDB database.

package mongodata

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"slices"
	"time"
)

type Server struct {
	Databases map[string]Database
}

type Database struct {
	Collections map[string]Collection
}

type Collection struct {
	Count int64
	Data  []bson.M
}

type Engine struct {
	Client *mongo.Client
	Server *Server
}

func (m *Engine) SetDatabases() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.Server.Databases[dbName] = Database{Collections: make(map[string]Collection)} // zero out the Data
	for _, collectionName := range collectionNames {
		coll := db.Collection(collectionName)
		c, err := coll.CountDocuments(context.Background(), bson.D{})
		if err != nil {
			return err
		}
		m.Server.Databases[dbName].Collections[collectionName] = Collection{Count: c}
	}

	return nil
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
