// mongodata handles the fetching and caching of data retrieved from the MongoDB database.

package mongodata

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"slices"
	"strings"
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

type selectedData struct {
	collectionName string
	databaseName   string
	documentId     string
}

type Engine struct {
	Client       *mongo.Client
	Server       *Server
	selectedData selectedData
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
		return fmt.Errorf("could not list collections for database %s: %v", dbName, err)
	}

	m.Server.Databases[dbName] = Database{Collections: make(map[string]Collection)} // zero out the Data
	for _, collectionName := range collectionNames {
		m.Server.Databases[dbName].Collections[collectionName] = Collection{}
	}

	return nil
}

// SetSelectedCollection Allows parent components to set what data will be displayed within this component.
func (m *Engine) SetSelectedCollection(collectionName, databaseName string) {
	m.selectedData = selectedData{
		collectionName: collectionName,
		databaseName:   databaseName,
	}
}

func (m *Engine) SetSelectedDocument(documentId string) {
	m.selectedData.documentId = documentId
}

func (m *Engine) IsDocumentSelected() bool {
	return m.selectedData.documentId != ""
}

func (m *Engine) GetSelectedDocument() (string, error) {
	if m.selectedData.databaseName == "" || m.selectedData.collectionName == "" || m.selectedData.documentId == "" {
		return "", fmt.Errorf("no document selected") // This should never happen if calling components are working correctly
	}

	db := m.Client.Database(m.selectedData.databaseName)
	coll := db.Collection(m.selectedData.collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	trimmedDoc := strings.TrimPrefix(m.selectedData.documentId, `ObjectID("`) // TODO look into how this is set and if it can be done better
	trimmedDoc = strings.TrimSuffix(trimmedDoc, `")`)

	objectId, err := bson.ObjectIDFromHex(trimmedDoc)
	if err != nil {
		return "", fmt.Errorf("invalid ObjectID %s: %v", trimmedDoc, err)
	}
	cur := coll.FindOne(ctx, bson.D{{"_id", objectId}})
	if cur.Err() != nil {
		return "", fmt.Errorf("failed to retrieve document with Id '%s': %w", m.selectedData.documentId, cur.Err())
	}
	data, err := cur.Raw()
	if err != nil {
		return "", err
	}

	parsedDoc, err := bson.MarshalExtJSON(data, false, false)
	if err != nil {
		return "", fmt.Errorf("could not parse document: %v", err)
	}

	return string(parsedDoc), nil
}

func (m *Engine) IsCollectionSelected() bool {
	return m.selectedData.collectionName != "" && m.selectedData.databaseName != ""
}

// QueryCollection fetches all the data from a given collection in a given database
func (m *Engine) QueryCollection(query bson.D) error {
	if m.selectedData.databaseName == "" || m.selectedData.collectionName == "" {
		return fmt.Errorf("no collection selected") // This should never happen
	}

	db := m.Client.Database(m.selectedData.databaseName)
	coll := db.Collection(m.selectedData.collectionName)
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

	curDb, ok := m.Server.Databases[m.selectedData.databaseName]
	if ok {
		curDb.Collections[m.selectedData.collectionName] = Collection{Data: data}
	} else {
		return fmt.Errorf("no database is set") // should never happen
	}
	return nil
}

// GetQueriedDocs returns the data that was last queried from the database for the selected database/collection
func (m *Engine) GetQueriedDocs() []bson.M {
	db, ok := m.Server.Databases[m.selectedData.databaseName]
	if ok {
		collection, ok := db.Collections[m.selectedData.collectionName]
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
