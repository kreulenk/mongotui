package mongoengine

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"slices"
)

// The methods contained in this file pertain to the displaying of data that has already
// been loaded in and cached to the Server struct

// GetDatabases will return a slice of all databases currently cached by mongotui
func (e *Engine) GetDatabases() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	databaseNames := make([]string, 0, len(e.server.databases))
	for name := range e.server.databases {
		databaseNames = append(databaseNames, name)
	}
	slices.Sort(databaseNames)
	return databaseNames
}

// GetSelectedCollections will return a slice of all collections currently cached by
// mongotui given the last database set by SetSelectedDatabase
func (e *Engine) GetSelectedCollections() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if db, ok := e.server.databases[e.selectedDb]; ok {
		return db.collections
	} else {
		return []string{}
	}
}

// GetDocumentSummaries will fetch a processed list of document summaries for each document
// These documents are currently being used to be displayed within the doclist component
func (e *Engine) GetDocumentSummaries() []docSummary {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.server.cachedDocSummaries
}

// GetSelectedDocumentMarshalled will return a marshalled byte slice of the document that was
// last selected via SetSelectedDocument
func (e *Engine) GetSelectedDocumentMarshalled() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	data := e.selectedDoc
	parsedDoc, err := bson.MarshalExtJSONIndent(data, false, false, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not parse document: %v", err)
	}

	return parsedDoc, nil
}

// GetSelectedDocument will return a reference to the bson of the highlighted doc
// last selected via SetSelectedDocument
func (e *Engine) GetSelectedDocument() *bson.M {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.selectedDoc
}

// GetQueriedDocs returns a slice of all documents cached by mongotui given the collection
// last selected via SetSelectedCollection
func (e *Engine) GetQueriedDocs() []*bson.M {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.server.cachedDocs
}
