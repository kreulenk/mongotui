// mongoengine handles the fetching and caching of data retrieved from the MongoDB database.

package mongoengine

import (
	"context"
	"fmt"
	"github.com/kreulenk/mongotui/internal/state"
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
type Engine struct {
	Client *mongo.Client
	Server *Server
	state  *state.TuiState
}

func New(client *mongo.Client, state *state.TuiState) *Engine {
	return &Engine{
		Client: client,
		Server: &Server{
			Databases: make(map[string]Database),
		},
		state: state,
	}
}

func (m *Engine) RefreshDatabases() error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	dbNames, err := m.Client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.Server.Databases = make(map[string]Database)
	for _, dbName := range dbNames {
		_, err := m.GetCollectionsOfDb(dbName)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCollectionsOfDb fetches the Collections for a given database along with the number of records in each collection
func (m *Engine) GetCollectionsOfDb(dbName string) ([]string, error) {
	_, ok := m.Server.Databases[dbName] // If we already have the cached data, don't fetch it again
	if ok {
		return getSortedCollectionsByName(m.Server.Databases[dbName].Collections), nil
	}

	db := m.Client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("could not list collections for database %s: %v", dbName, err)
	}

	m.Server.Databases[dbName] = Database{Collections: make(map[string]Collection)} // zero out the Data
	for _, collectionName := range collectionNames {
		m.Server.Databases[dbName].Collections[collectionName] = Collection{}
	}

	return getSortedCollectionsByName(m.Server.Databases[dbName].Collections), nil
}

func (m *Engine) ClearCachedData() {
	m.Server = &Server{
		Databases: make(map[string]Database),
	}
}

// getSortedCollectionsByName returns a slice of collection names sorted alphabetically
func getSortedCollectionsByName(collections map[string]Collection) []string {
	collectionNames := make([]string, 0, len(collections))
	for name := range collections {
		collectionNames = append(collectionNames, name)
	}
	slices.Sort(collectionNames)
	return collectionNames
}

func (m *Engine) GetSelectedDocument() ([]byte, error) {
	dbName := m.state.MainViewState.DbColTableState.GetSelectedDbName()
	colName := m.state.MainViewState.DbColTableState.GetSelectedCollectionName()
	docIndex := m.state.MainViewState.DocListState.GetSelectedDocumentIndex()

	data := m.Server.Databases[dbName].Collections[colName].Data[docIndex]
	parsedDoc, err := bson.MarshalExtJSONIndent(data, false, false, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not parse document: %v", err)
	}

	return parsedDoc, nil
}

// QueryCollection fetches all the data from a given collection in a given database
func (m *Engine) QueryCollection(query bson.D) error {
	db := m.Client.Database(m.state.MainViewState.DbColTableState.GetSelectedDbName())
	coll := db.Collection(m.state.MainViewState.DbColTableState.GetSelectedCollectionName())
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

	curDb, ok := m.Server.Databases[m.state.MainViewState.DbColTableState.GetSelectedDbName()]
	if ok {
		curDb.Collections[m.state.MainViewState.DbColTableState.GetSelectedCollectionName()] = Collection{Data: data}
	} else {
		return fmt.Errorf("no database is set") // should never happen
	}
	return nil
}

func (m *Engine) DropDatabase(databaseName string) error {
	db := m.Client.Database(databaseName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	return db.Drop(ctx)
}

func (m *Engine) DropCollection(databaseName, collectionName string) error {
	db := m.Client.Database(databaseName)
	col := db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	return col.Drop(ctx)
}

func (m *Engine) UpdateDocument(oldDoc, newDoc []byte) error {
	var oldDocBson bson.M
	if err := bson.UnmarshalExtJSON(oldDoc, false, &oldDocBson); err != nil {
		return fmt.Errorf("failed to parse the original document needed for the replacement: %w", err)
	}
	var newDocBson bson.M
	if err := bson.UnmarshalExtJSON(newDoc, false, &newDocBson); err != nil {
		return fmt.Errorf("failed to parse the new document needed for the replacement: %w", err)
	}

	db := m.Client.Database(m.state.MainViewState.DbColTableState.GetSelectedDbName())
	coll := db.Collection(m.state.MainViewState.DbColTableState.GetSelectedCollectionName())
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	res := coll.FindOneAndReplace(ctx, oldDocBson, newDocBson)
	if res.Err() != nil {
		return fmt.Errorf("failed to update document: %w", res.Err())
	}
	return nil
}

// GetQueriedDocs returns the data that was last queried from the database for the selected database/collection
func (m *Engine) GetQueriedDocs() []bson.M {
	db, ok := m.Server.Databases[m.state.MainViewState.DbColTableState.GetSelectedDbName()]
	if ok {
		collection, ok := db.Collections[m.state.MainViewState.DbColTableState.GetSelectedCollectionName()]
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
