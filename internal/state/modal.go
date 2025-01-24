package state

type ModalState struct {
	err                                    error
	requestedDatabaseModalDeletionPrompt   bool // sent from dbcoltable to request that the DbColTableState databaseName be deleted
	requestedCollectionModalDeletionPrompt bool // sent from dbcoltable to request that the DbColTableState collectionName be deleted
}

func (m *ModalState) SetError(err error) {
	m.err = err
}

func (m *ModalState) GetError() error {
	return m.err
}

func (m *ModalState) ClearError() {
	m.err = nil
}

// RequestDatabaseModalDeletionPrompt can be called to signal to the mainview component
// that it should display the modal component with a prompt to confirm that the
// selected database signaled by DbColTableState databaseName should be deleted
func (m *ModalState) RequestDatabaseModalDeletionPrompt() {
	m.requestedDatabaseModalDeletionPrompt = true
}

// RequestCollectionModalDeletionPrompt can be called to signal to the mainview component
// that it should display the modal component with a prompt to confirm that the
// selected collection signaled by DbColTableState databaseName and DbColTableState collectionName should be deleted
func (m *ModalState) RequestCollectionModalDeletionPrompt() {
	m.requestedCollectionModalDeletionPrompt = true
}

// ResetDatabaseDeletionPromptRequest would be called by the modal component after a user has
// confirmed whether or not they would like to actually delete a specific database
func (m *ModalState) ResetDatabaseDeletionPromptRequest() {
	m.requestedDatabaseModalDeletionPrompt = false
}

// ResetCollectionDeletionPromptRequest would be called by the modal component after a user has
// confirmed whether or not they would like to actually delete a specific collection
func (m *ModalState) ResetCollectionDeletionPromptRequest() {
	m.requestedCollectionModalDeletionPrompt = false
}

// IsDatabaseDeletionRequested can be used to check whether the DbColTable has requested
// that the modal component be displayed to confirm whether or not the selected DbColTableState databaseName
// should be deleted
func (m *ModalState) IsDatabaseDeletionRequested() bool {
	return m.requestedDatabaseModalDeletionPrompt
}

// IsCollectionDeletionRequested can be used to check whether the DbColTable has requested
// that the modal component be displayed to confirm whether or not the selected DbColTableState collectionName
// should be deleted
func (m *ModalState) IsCollectionDeletionRequested() bool {
	return m.requestedCollectionModalDeletionPrompt
}
