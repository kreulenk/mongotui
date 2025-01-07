// mongodata handles the fetching and caching of data retrieved from the MongoDB database.

package mongodata

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"slices"
	"time"
)

const timeout = 5 * time.Second

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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

// GetData fetches all the data from a given collection in a given database
// TODO: add pagination
func (m *Engine) GetData(dbName, collectionName string) ([]bson.M, error) {
	if dbName == "" || collectionName == "" {
		return nil, nil
	}

	db := m.Client.Database(dbName)
	coll := db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cur, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var data []bson.M
	if err = cur.All(ctx, &data); err != nil {
		return nil, err
	}

	return data, nil
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
