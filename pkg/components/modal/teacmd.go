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
Database Delete Modal
************************
*/

type DbDeleteModalMsg struct {
	dbName string
}

func DisplayDatabaseDeleteModal(dbName string) tea.Cmd {
	return func() tea.Msg {
		return DbDeleteModalMsg{
			dbName: dbName,
		}
	}
}

type ExecDbDelete struct {
	DbName string
}

func execDatabaseDelete(dbName string) tea.Cmd {
	return func() tea.Msg {
		return ExecDbDelete{
			DbName: dbName,
		}
	}
}

/*
************************
Collection Delete Modal
************************
*/

type ColDeleteModalMsg struct {
	dbName         string
	collectionName string
}

func DisplayCollectionDeleteModal(dbName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return ColDeleteModalMsg{
			dbName:         dbName,
			collectionName: collectionName,
		}
	}
}

// ExecColDelete will be sent down to the dbcoltable to actually delete a collection after a modal confirmation as
// well as tell the table to refresh its data
type ExecColDelete struct {
	DbName         string
	CollectionName string
}

func execCollectionDelete(dbName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return ExecColDelete{
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
