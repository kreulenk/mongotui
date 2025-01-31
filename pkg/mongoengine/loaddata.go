package mongoengine

// The methods contained within this file pertain to loading in data from the actual mongo database
// and caching it for future use

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"slices"
	"strings"
)

func (m *Engine) RefreshDbAndCollections() tea.Cmd {
	m.Server = &Server{ // Clear all cached data
		Databases: make(map[string]Database),
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	dbNames, err := m.Client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return modal.DisplayErrorModal(err)
	}

	m.Server.Databases = make(map[string]Database)
	for _, dbName := range dbNames {
		if err := m.fetchCollectionsPerDb(dbName); err != nil {
			return modal.DisplayErrorModal(err)
		}
	}
	return nil
}

// fetchCollectionsPerDb will add all collections to a database entry within the mongoengine Server struct
func (m *Engine) fetchCollectionsPerDb(dbName string) error {
	db := m.Client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	slices.Sort(collectionNames)
	if err != nil {
		return fmt.Errorf("could not list collections for database %s: %v", dbName, err)
	}
	m.Server.Databases[dbName] = Database{Collections: make(map[string]Collection), cachedSortedCollectionNames: collectionNames} // zero out the Documents
	for _, collectionName := range collectionNames {
		m.Server.Databases[dbName].Collections[collectionName] = Collection{}
	}
	return nil
}

// QueryCollection fetches all the data from a given collection in a given database
func (m *Engine) QueryCollection(query bson.D) tea.Cmd {
	db := m.Client.Database(m.selectedDb)
	coll := db.Collection(m.selectedCollection)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	cur, err := coll.Find(ctx, query)
	if err != nil {
		return modal.DisplayErrorModal(err)
	}

	var data []*bson.M
	if err = cur.All(ctx, &data); err != nil {
		return modal.DisplayErrorModal(err)
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
