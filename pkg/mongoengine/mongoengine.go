// mongoengine handles the fetching and caching of data retrieved from the MongoDB database.

package mongoengine

import (
	"context"
	"fmt"
	"github.com/kreulenk/mongotui/internal/state"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log/slog"
	"slices"
	"strings"
	"time"
)

const Timeout = 5 * time.Second

type Server struct {
	Databases           map[string]Database
	cachedSortedDbNames []string
}

type Database struct {
	Collections                 map[string]Collection
	cachedSortedCollectionNames []string
}

type Collection struct {
	Documents          []*bson.M
	cachedDocSummaries []DocSummary // Used to display a little information about each doc in the doclist component
}

type DocSummary []FieldSummary // Used to display a little information about each doc in the doclist component

type FieldSummary struct {
	Name  string
	Type  string // TODO restrict to a set of types
	Value string
}

type Engine struct {
	Client *mongo.Client
	Server *Server
	state  *state.TuiState

	selectedDb         string
	selectedCollection string
	selectedDoc        *bson.M
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

func (m *Engine) SetSelectedCollection(d, c string) {
	m.selectedDb = d
	m.selectedCollection = c
}

func (m *Engine) SetSelectedDatabase(d string) {
	m.selectedDb = d
	m.selectedCollection = ""
}

func (m *Engine) SetSelectedDocument(d *bson.M) {
	m.selectedDoc = d
}

func (m *Engine) GetDatabases() []string {
	if m.Server.cachedSortedDbNames == nil {
		databaseNames := make([]string, 0, len(m.Server.Databases))
		for name := range m.Server.Databases {
			databaseNames = append(databaseNames, name)
		}
		slices.Sort(databaseNames)
		m.Server.cachedSortedDbNames = databaseNames
		return databaseNames
	}

	return m.Server.cachedSortedDbNames
}

func (m *Engine) GetSelectedCollections() []string {
	if db, ok := m.Server.Databases[m.selectedDb]; ok {
		return db.cachedSortedCollectionNames
	} else {
		slog.Info("GetSelectedCollections called with no selectedDb")
		return []string{}
	}
}

// GetDocumentSummaries will fetch a processed list of document summaries for each document
// These documents are currently being used to be displayed within the doclist component
// TODO add cacheing to this
func (m *Engine) GetDocumentSummaries() []DocSummary {
	if cachedCol, ok := m.Server.Databases[m.selectedDb].Collections[m.selectedCollection]; ok {
		return cachedCol.cachedDocSummaries
	} else {
		return []DocSummary{} // Return empty arr until data has loaded
	}
}

func getFieldType(value interface{}) string {
	switch v := value.(type) {
	case bson.M, bson.D:
		return "Object"
	case bson.A:
		return "Array"
	case nil:
		return "Null"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// TODO consider removing this as it is not needed with m.GetDatbases already existing
func (m *Engine) GetDatabaseByIndex(index int) string {
	m.GetDatabases()
	return m.Server.cachedSortedDbNames[index]
}

func (m *Engine) RefreshDbAndCollections() error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	dbNames, err := m.Client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.Server.Databases = make(map[string]Database)
	for _, dbName := range dbNames {
		_, err := m.FetchCollectionsPerDb(dbName)
		if err != nil {
			return err
		}
	}

	return nil
}

// FetchCollectionsPerDb fetches the Collections for a given database along with the number of records in each collection
func (m *Engine) FetchCollectionsPerDb(dbName string) ([]string, error) {
	db := m.Client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	slices.Sort(collectionNames)
	if err != nil {
		return nil, fmt.Errorf("could not list collections for database %s: %v", dbName, err)
	}

	m.Server.Databases[dbName] = Database{Collections: make(map[string]Collection), cachedSortedCollectionNames: collectionNames} // zero out the Documents
	for _, collectionName := range collectionNames {
		m.Server.Databases[dbName].Collections[collectionName] = Collection{}
	}

	return collectionNames, nil
}

func (m *Engine) ClearCachedData() {
	m.Server = &Server{
		Databases: make(map[string]Database),
	}
}

// TODO verify if this pointer should be dereferenced
func (m *Engine) GetSelectedDocumentMarshalled() ([]byte, error) {
	data := m.selectedDoc
	parsedDoc, err := bson.MarshalExtJSONIndent(data, false, false, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not parse document: %v", err)
	}

	return parsedDoc, nil
}

// QueryCollection fetches all the data from a given collection in a given database
func (m *Engine) QueryCollection(query bson.D) error {
	db := m.Client.Database(m.selectedDb)
	coll := db.Collection(m.selectedCollection)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	cur, err := coll.Find(ctx, query)
	if err != nil {
		return err
	}

	var data []*bson.M
	if err = cur.All(ctx, &data); err != nil {
		return err
	}

	// Create doc summary cache
	var newDocsSummaries []DocSummary
	for _, doc := range data {
		var row DocSummary
		for k, v := range *doc {
			val := fmt.Sprintf("%v", v)
			fType := getFieldType(v)
			if fType == "Object" || fType == "Array" { // TODO restrict to a set of types
				val = fType
			}
			row = append(row, FieldSummary{
				Name:  k,
				Type:  fType,
				Value: val,
			})
		}
		slices.SortFunc(row, func(i, j FieldSummary) int { // Sort the fields being returned
			return strings.Compare(i.Name, j.Name)
		})
		newDocsSummaries = append(newDocsSummaries, row)
	}

	m.Server.Databases[m.selectedDb].Collections[m.selectedCollection] = Collection{Documents: data, cachedDocSummaries: newDocsSummaries}
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

// UpdateDocument will find and replace a given oldDoc with a newDoc within the db/collection
// that was selected using the SetSelectedCollection method
func (m *Engine) UpdateDocument(oldDoc, newDoc []byte) error {
	var oldDocBson bson.M
	if err := bson.UnmarshalExtJSON(oldDoc, false, &oldDocBson); err != nil {
		return fmt.Errorf("failed to parse the original document needed for the replacement: %w", err)
	}
	var newDocBson bson.M
	if err := bson.UnmarshalExtJSON(newDoc, false, &newDocBson); err != nil {
		return fmt.Errorf("failed to parse the new document needed for the replacement: %w", err)
	}

	db := m.Client.Database(m.selectedDb)
	coll := db.Collection(m.selectedCollection)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	res := coll.FindOneAndReplace(ctx, oldDocBson, newDocBson)
	if res.Err() != nil {
		return fmt.Errorf("failed to update document: %w", res.Err())
	}
	return nil
}

// GetQueriedDocs returns the data that was last queried from the database for the selected database/collection
func (m *Engine) GetQueriedDocs() []*bson.M {
	db, ok := m.Server.Databases[m.selectedDb]
	if ok {
		collection, ok := db.Collections[m.selectedCollection]
		if ok {
			return collection.Documents
		}
	}
	return nil
}
