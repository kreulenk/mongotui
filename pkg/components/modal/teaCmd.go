package modal

import tea "github.com/charmbracelet/bubbletea"

type ErrModalMsg struct {
	err error
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
		return ErrModalMsg{err: err}
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

func execCollectionDelete(dbName, collectionName string) tea.Cmd {
	return func() tea.Msg {
		return ExecColDelete{
			DbName:         dbName,
			CollectionName: collectionName,
		}
	}
}
