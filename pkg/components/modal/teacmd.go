package modal

import tea "github.com/charmbracelet/bubbletea"

type ErrModalMsg struct {
	Err error
}

type DbDeleteModalMsg struct {
	dbName string
}

type ExecDbDelete struct {
	DbName string
}

type ColDeleteModalMsg struct {
	dbName         string
	collectionName string
}

// ExecColDelete will be sent down to the dbcoltable to actually delete a collection after a modal confirmation as
// well as tell the table to refresh its data
type ExecColDelete struct {
	DbName         string
	CollectionName string
}

func DisplayErrorModal(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrModalMsg{Err: err}
	}
}

func DisplayDatabaseDeleteModal(dbName string) tea.Cmd {
	return func() tea.Msg {
		return DbDeleteModalMsg{
			dbName: dbName,
		}
	}
}

func DisplayCollectionDeleteModal(dbName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return ColDeleteModalMsg{
			dbName:         dbName,
			collectionName: collectionName,
		}
	}
}

func execDatabaseDelete(dbName string) tea.Cmd {
	return func() tea.Msg {
		return ExecDbDelete{
			DbName: dbName,
		}
	}
}

func execCollectionDelete(dbName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return ExecColDelete{
			DbName:         dbName,
			CollectionName: collectionName,
		}
	}
}
