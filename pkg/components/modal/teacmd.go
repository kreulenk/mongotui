package modal

import (
	tea "github.com/charmbracelet/bubbletea"
	"go.mongodb.org/mongo-driver/v2/bson"
)

/*
************************
Error Modal
************************
*/

type ErrModalMsg struct {
	Err error
}

func DisplayErrorModal(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrModalMsg{Err: err}
	}
}

/*
************************
Insert Collection Modal
************************
*/

type DbCollInsertModalMsg struct {
	databaseName string
}

// DisplayDbCollInsertModal is called from dbcoltable. Is the user is already cursored into a collection
// then a databaseName should be specified. If not, an empty string should be passed
func DisplayDbCollInsertModal(databaseName string) tea.Cmd {
	return func() tea.Msg {
		return DbCollInsertModalMsg{databaseName: databaseName}
	}
}

type ExecDbCollInsert struct {
	DatabaseName   string
	CollectionName string
}

func execDbCollectionInsert(databaseName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return ExecDbCollInsert{DatabaseName: databaseName, CollectionName: collectionName}
	}
}

/*
************************
Database Delete Modal
************************
*/

type DbDropModalMsg struct {
	dbName string
}

func DisplayDatabaseDropModal(dbName string) tea.Cmd {
	return func() tea.Msg {
		return DbDropModalMsg{
			dbName: dbName,
		}
	}
}

type ExecDbDrop struct {
	DbName string
}

func execDatabaseDrop(dbName string) tea.Cmd {
	return func() tea.Msg {
		return ExecDbDrop{
			DbName: dbName,
		}
	}
}

/*
************************
Collection Delete Modal
************************
*/

type CollDropModalMsg struct {
	dbName         string
	collectionName string
}

func DisplayCollectionDropModal(dbName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return CollDropModalMsg{
			dbName:         dbName,
			collectionName: collectionName,
		}
	}
}

// ExecCollDrop will be sent down to the dbcoltable to actually drop a collection after a modal confirmation as
// well as tell the table to refresh its data
type ExecCollDrop struct {
	DbName         string
	CollectionName string
}

func execCollectionDrop(dbName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return ExecCollDrop{
			DbName:         dbName,
			CollectionName: collectionName,
		}
	}
}

/*
************************
Document Delete Modal
************************
*/

type DocDeleteModalMsg struct {
	doc *bson.M
}

func DisplayDocDeleteModal(doc *bson.M) tea.Cmd {
	return func() tea.Msg {
		return DocDeleteModalMsg{
			doc: doc,
		}
	}
}

type ExecDocDelete struct {
	Doc *bson.M
}

func execDocDelete(doc *bson.M) tea.Cmd {
	return func() tea.Msg {
		return ExecDocDelete{Doc: doc}
	}
}

/*
************************
Document Insert Modal
************************
*/

type DocInsertModalMsg struct {
	doc bson.M
}

func DisplayDocInsertModal(doc bson.M) tea.Cmd {
	return func() tea.Msg {
		return DocInsertModalMsg{
			doc: doc,
		}
	}
}

type ExecDocInsert struct {
	Doc bson.M
}

func execDocInsert(doc bson.M) tea.Cmd {
	return func() tea.Msg {
		return ExecDocInsert{Doc: doc}
	}
}

/*
************************
Document Edit Modal
************************
*/

type DocEditModalMsg struct {
	oldDoc bson.M
	newDoc bson.M
}

func DisplayDocEditModal(oldDoc, newDoc bson.M) tea.Cmd {
	return func() tea.Msg {
		return DocEditModalMsg{
			oldDoc: oldDoc,
			newDoc: newDoc,
		}
	}
}

type ExecDocEdit struct {
	OldDoc bson.M
	NewDoc bson.M
}

func execDocEdit(oldDoc, newDoc bson.M) tea.Cmd {
	return func() tea.Msg {
		return ExecDocEdit{OldDoc: oldDoc, NewDoc: newDoc}
	}
}
