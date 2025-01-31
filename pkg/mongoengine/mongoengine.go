// mongoengine handles the fetching and caching of data retrieved from the MongoDB database.

package mongoengine

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

	selectedDb         string
	selectedCollection string
	selectedDoc        *bson.M
}

func New(client *mongo.Client) *Engine {
	return &Engine{
		Client: client,
		Server: &Server{
			Databases: make(map[string]Database),
		},
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

func (m *Engine) DropDatabase(databaseName string) tea.Cmd {
	db := m.Client.Database(databaseName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	if err := db.Drop(ctx); err != nil {
		return modal.DisplayErrorModal(err)
	}
	return m.RefreshDbAndCollections()
}

func (m *Engine) DropCollection(databaseName, collectionName string) tea.Cmd {
	db := m.Client.Database(databaseName)
	col := db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	if err := col.Drop(ctx); err != nil {
		return modal.DisplayErrorModal(err)
	}

	if len(m.GetSelectedCollections()) == 1 { // If we just dropped the last collection, reset selectedCollection
		m.SetSelectedDatabase(databaseName)
	}
	return m.RefreshDbAndCollections()
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
