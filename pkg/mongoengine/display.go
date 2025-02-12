package mongoengine

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"slices"
)

// The methods contained in this file pertain to the displaying of data that has already
// been loaded in and cached to the Server struct

// GetDatabases will return a slice of all databases currently cached by mongotui
func (m *Engine) GetDatabases() []string {
	databaseNames := make([]string, 0, len(m.Server.Databases))
	for name := range m.Server.Databases {
		databaseNames = append(databaseNames, name)
	}
	slices.Sort(databaseNames)
	return databaseNames
}

// GetSelectedCollections will return a slice of all of the collections currently cached by
// mongotui given the last database set by SetSelectedDatabase
func (m *Engine) GetSelectedCollections() []string {
	if db, ok := m.Server.Databases[m.selectedDb]; ok {
		return db.collections
	} else {
		return []string{}
	}
}

// GetDocumentSummaries will fetch a processed list of document summaries for each document
// These documents are currently being used to be displayed within the doclist component
func (m *Engine) GetDocumentSummaries() []DocSummary {
	return m.Server.cachedDocSummaries
}

// GetSelectedDocumentMarshalled will return a marshalled byte slice of the document that was
// last selected via SetSelectedDocument
func (m *Engine) GetSelectedDocumentMarshalled() ([]byte, error) {
	data := m.selectedDoc
	parsedDoc, err := bson.MarshalExtJSONIndent(data, false, false, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not parse document: %v", err)
	}

	return parsedDoc, nil
}

// GetSelectedDocument will return a reference to the bson of the highlighted doc
// last selected via SetSelectedDocument
func (m *Engine) GetSelectedDocument() *bson.M {
	return m.selectedDoc
}

// GetQueriedDocs returns a slice of all documents cached by mongotui given the collection
// last selected via SetSelectedCollection
func (m *Engine) GetQueriedDocs() []*bson.M {
	return m.Server.cachedDocs
}
