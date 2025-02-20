package mongoengine

// The methods contained within this file pertain to loading in data from the actual mongo database
// and caching it for future use

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"slices"
	"strings"
)

func (e *Engine) RefreshDbAndCollections() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.server = &server{ // Clear all cached data
		databases: make(map[string]database),
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	dbNames, err := e.Client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	e.server.databases = make(map[string]database)
	for _, dbName := range dbNames {
		if err := e.fetchCollectionsPerDb(dbName); err != nil {
			return err
		}
	}
	e.server.cachedDocs = nil
	e.server.cachedDocSummaries = nil
	e.DocCount = 0
	return nil
}

// fetchCollectionsPerDb will add all collections to a database entry within the mongoengine Server struct
func (e *Engine) fetchCollectionsPerDb(dbName string) error {
	db := e.Client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("could not list collections for database %s: %v", dbName, err)
	}
	slices.Sort(collectionNames)
	e.server.databases[dbName] = database{collections: collectionNames} // zero out the Documents
	return nil
}

// QueryCollection fetches all the data from a given collection in a given database given a particular query
func (e *Engine) QueryCollection(query bson.D) tea.Cmd {
	return func() tea.Msg {
		e.mu.Lock()
		defer e.mu.Unlock()

		e.Skip = 0
		count, err := e.getDocCount(query)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		e.DocCount = count
		err = e.executeQuery(query)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}

		e.lastExecutedQuery = query
		return RedrawMessage{}
	}
}

// RerunLastCollectionQuery will rerun the last query that was just run against the database. This is useful
// after doc edits or pagination
func (e *Engine) RerunLastCollectionQuery() tea.Cmd {
	return func() tea.Msg {
		e.mu.Lock()
		defer e.mu.Unlock()

		err := e.executeQuery(e.lastExecutedQuery)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		return RedrawMessage{}
	}
}

func (e *Engine) NextPage() tea.Cmd {
	if e.Skip+Limit < e.DocCount {
		e.Skip += Limit
		return e.RerunLastCollectionQuery()
	} else {
		return modal.DisplayErrorModal(fmt.Errorf("already on last document page"))
	}
}

func (e *Engine) PreviousPage() tea.Cmd {
	if e.Skip > 0 {
		e.Skip -= Limit
		return e.RerunLastCollectionQuery()
	} else {
		return modal.DisplayErrorModal(fmt.Errorf("already on first document page"))
	}
}

func (e *Engine) getDocCount(query bson.D) (int64, error) {
	db := e.Client.Database(e.selectedDb)
	coll := db.Collection(e.selectedCollection)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	return coll.CountDocuments(ctx, query)
}

// executeQuery is used internally to handle an initial query sent by QueryCollection as well as the nextPage and
// previousPage pagination functions that use the m.lastExecutedQuery variable
func (e *Engine) executeQuery(query bson.D) error {
	db := e.Client.Database(e.selectedDb)
	coll := db.Collection(e.selectedCollection)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	findOptions := options.Find().SetSkip(e.Skip).SetLimit(int64(Limit))
	cur, err := coll.Find(ctx, query, findOptions)
	if err != nil {
		return err
	}

	var data []*bson.M
	if err = cur.All(ctx, &data); err != nil {
		return err
	}

	// Create doc summary cache
	var newDocsSummaries []docSummary
	for _, doc := range data {
		var row docSummary
		for k, v := range *doc {
			val := fmt.Sprintf("%v", v)
			fType := getFieldType(v)
			if fType == "Object" || fType == "Array" { // TODO restrict to a set of types
				val = fType
			}
			row = append(row, fieldSummary{
				Name:  k,
				Type:  fType,
				Value: val,
			})
		}
		slices.SortFunc(row, func(i, j fieldSummary) int { // Sort the fields being returned
			return strings.Compare(i.Name, j.Name)
		})
		newDocsSummaries = append(newDocsSummaries, row)
	}

	e.server.cachedDocSummaries = newDocsSummaries
	e.server.cachedDocs = data
	return nil
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
