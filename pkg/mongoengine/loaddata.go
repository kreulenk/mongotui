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

func (m *Engine) RefreshDbAndCollections() error {
	m.Server = &Server{ // Clear all cached data
		Databases: make(map[string]Database),
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	dbNames, err := m.Client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	m.Server.Databases = make(map[string]Database)
	for _, dbName := range dbNames {
		if err := m.fetchCollectionsPerDb(dbName); err != nil {
			return err
		}
	}
	m.Server.cachedDocs = nil
	m.Server.cachedDocSummaries = nil
	m.DocCount = 0
	return nil
}

// fetchCollectionsPerDb will add all collections to a database entry within the mongoengine Server struct
func (m *Engine) fetchCollectionsPerDb(dbName string) error {
	db := m.Client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("could not list collections for database %s: %v", dbName, err)
	}
	slices.Sort(collectionNames)
	m.Server.Databases[dbName] = Database{collections: collectionNames} // zero out the Documents
	return nil
}

// QueryCollection fetches all the data from a given collection in a given database given a particular query
func (m *Engine) QueryCollection(query bson.D) tea.Cmd {
	return func() tea.Msg {
		m.Skip = 0
		count, err := m.getDocCount(query)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		m.DocCount = count
		err = m.executeQuery(query)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}

		m.lastExecutedQuery = query
		return RedrawMessage{}
	}
}

// RerunLastCollectionQuery will rerun the last query that was just run against the database. This is useful
// after doc edits or pagination
func (m *Engine) RerunLastCollectionQuery() tea.Cmd {
	return func() tea.Msg {
		err := m.executeQuery(m.lastExecutedQuery)
		if err != nil {
			return modal.ErrModalMsg{Err: err}
		}
		return RedrawMessage{}
	}
}

func (m *Engine) NextPage() tea.Cmd {
	if m.Skip+Limit < m.DocCount {
		m.Skip += Limit
		return m.RerunLastCollectionQuery()
	} else {
		return modal.DisplayErrorModal(fmt.Errorf("already on last document page"))
	}
}

func (m *Engine) PreviousPage() tea.Cmd {
	if m.Skip > 0 {
		m.Skip -= Limit
		return m.RerunLastCollectionQuery()
	} else {
		return modal.DisplayErrorModal(fmt.Errorf("already on first document page"))
	}
}

func (m *Engine) getDocCount(query bson.D) (int64, error) {
	db := m.Client.Database(m.selectedDb)
	coll := db.Collection(m.selectedCollection)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	return coll.CountDocuments(ctx, query)
}

// executeQuery is used internally to handle an initial query sent by QueryCollection as well as the nextPage and
// previousPage pagination functions that use the m.lastExecutedQuery variable
func (m *Engine) executeQuery(query bson.D) error {
	db := m.Client.Database(m.selectedDb)
	coll := db.Collection(m.selectedCollection)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	findOptions := options.Find().SetSkip(m.Skip).SetLimit(int64(Limit))
	cur, err := coll.Find(ctx, query, findOptions)
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

	m.Server.cachedDocSummaries = newDocsSummaries
	m.Server.cachedDocs = data
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
