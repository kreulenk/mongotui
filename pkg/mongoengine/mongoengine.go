// mongoengine handles the fetching and caching of data retrieved from the MongoDB database.

package mongoengine

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"sync"
	"time"
)

const Timeout = 10 * time.Second
const Limit = 25 // Page size for doclist component

type server struct {
	databases map[string]database

	// Info about docs being displayed in doclist component
	cachedDocSummaries []docSummary
	cachedDocs         []*bson.M
}

type database struct {
	collections []string
}

type docSummary []fieldSummary // Used to display information about each doc in the doclist component

type fieldSummary struct {
	Name  string
	Type  string // TODO restrict to a set of types
	Value string
}

type Engine struct {
	Client *mongo.Client
	server *server

	selectedDb         string
	selectedCollection string
	selectedDoc        *bson.M

	lastExecutedQuery bson.D // Used to refresh db after deletion operation and for pagination
	Skip              int64  // Used for pagination when querying docs
	DocCount          int64  // Used for pagination

	mu sync.RWMutex // bubbletea sends updates in go routines concurrently
}

func New(client *mongo.Client) *Engine {
	return &Engine{
		Client: client,
		server: &server{
			databases: make(map[string]database),
		},
	}
}

func (e *Engine) SetSelectedCollection(d, c string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.selectedDb = d
	e.selectedCollection = c
}

func (e *Engine) SetSelectedDatabase(d string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.selectedDb = d
	e.selectedCollection = ""
}

func (e *Engine) SetSelectedDocument(d *bson.M) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.selectedDoc = d
}

func (e *Engine) DropDatabase(databaseName string) tea.Cmd {
	return func() tea.Msg {
		db := e.Client.Database(databaseName)
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()

		if err := db.Drop(ctx); err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		if err := e.RefreshDbAndCollections(); err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		return RedrawMessage{}
	}
}

func (e *Engine) DropCollection(databaseName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		db := e.Client.Database(databaseName)
		col := db.Collection(collectionName)
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()

		if err := col.Drop(ctx); err != nil {
			return modal.ErrModalMsg{Err: err}
		}

		if len(e.GetSelectedCollections()) == 1 { // If we just dropped the last collection, reset selectedCollection
			e.SetSelectedDatabase(databaseName)
		}
		if err := e.RefreshDbAndCollections(); err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		return RedrawMessage{}
	}
}

// DropDocument will drop a document from the collection that was selected using SetSelectedCollection
// We do not use _id as not every doc will have one so we match the entire doc
func (e *Engine) DropDocument(doc *bson.M) tea.Cmd {
	return func() tea.Msg {
		db := e.Client.Database(e.selectedDb)
		coll := db.Collection(e.selectedCollection)
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()

		res, err := coll.DeleteOne(ctx, doc)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		if res.DeletedCount == 0 {
			return modal.ErrModalMsg{Err: fmt.Errorf("no document was deleted")}
		}
		return e.RerunLastCollectionQuery()() // Double call as we are already calling it in a query
	}
}

// UpdateDocument will find and replace a given oldDoc with a newDoc within the db/collection
// that was selected using the SetSelectedCollection method
// We do not use _id as not every doc will have one so we match the entire doc
func (e *Engine) UpdateDocument(oldDoc, newDoc []byte) error {
	var oldDocBson bson.M
	if err := bson.UnmarshalExtJSON(oldDoc, false, &oldDocBson); err != nil {
		return fmt.Errorf("failed to parse the original document needed for the replacement: %w", err)
	}
	var newDocBson bson.M
	if err := bson.UnmarshalExtJSON(newDoc, false, &newDocBson); err != nil {
		return fmt.Errorf("failed to parse the new document needed for the replacement: %w", err)
	}

	db := e.Client.Database(e.selectedDb)
	coll := db.Collection(e.selectedCollection)
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
