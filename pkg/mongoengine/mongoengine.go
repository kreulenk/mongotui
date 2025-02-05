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

const Timeout = 15 * time.Second

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

	lastExecutedQuery bson.D // Used to refresh db after deletion operation
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
	return func() tea.Msg {
		db := m.Client.Database(databaseName)
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()
		if err := db.Drop(ctx); err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		if err := m.RefreshDbAndCollections(); err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		return RedrawMessage{}
	}
}

func (m *Engine) DropCollection(databaseName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		db := m.Client.Database(databaseName)
		col := db.Collection(collectionName)
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()
		if err := col.Drop(ctx); err != nil {
			return modal.ErrModalMsg{Err: err}
		}

		if len(m.GetSelectedCollections()) == 1 { // If we just dropped the last collection, reset selectedCollection
			m.SetSelectedDatabase(databaseName)
		}
		if err := m.RefreshDbAndCollections(); err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		return RedrawMessage{}
	}
}

// DropDocument will drop a document from the collection that was selected using SetSelectedCollection
// We do not use _id as not every doc will have one so we match the entire doc
// currentQuery is used to requery the database after the drop has been performed
func (m *Engine) DropDocument(doc *bson.M) tea.Cmd {
	return func() tea.Msg {
		db := m.Client.Database(m.selectedDb)
		coll := db.Collection(m.selectedCollection)
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()

		res, err := coll.DeleteOne(ctx, doc)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		if res.DeletedCount == 0 {
			return modal.ErrModalMsg{Err: fmt.Errorf("no document was deleted")}
		}
		return m.QueryCollection(m.lastExecutedQuery)() // Double call as we are already calling it in a query
	}
}

// UpdateDocument will find and replace a given oldDoc with a newDoc within the db/collection
// that was selected using the SetSelectedCollection method
// We do not use _id as not every doc will have one so we match the entire doc
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

// RedrawMessage is used to trigger a bubbletea update so that the components refresh their View functions
// whenever the underlying data within mongoengine has updated
type RedrawMessage struct{}
