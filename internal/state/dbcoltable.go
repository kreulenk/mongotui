package state

type DbColTableState struct {
	databaseName   string
	collectionName string

	databaseDeletionSuccess   bool
	collectionDeletionSuccess bool
}

func (d *DbColTableState) IsCollectionSelected() bool {
	return d.collectionName != "" && d.databaseName != ""
}

func (d *DbColTableState) ClearCollectionSelection() {
	d.collectionName = ""
	d.databaseName = ""
}

// GetSelectedDbName is used by the mongoengine package to determine what database
// has been selected within the dbcoltable component
func (d *DbColTableState) GetSelectedDbName() string {
	return d.databaseName
}

// GetSelectedCollectionName is used by the mongoengine package to determine what collection
// has been selected within the dbcoltable component
func (d *DbColTableState) GetSelectedCollectionName() string {
	return d.collectionName
}

// SetSelectedCollection is called by the dbcoltable component to signal to other components
// information about what database/collection are selected in the UI
func (d *DbColTableState) SetSelectedCollection(dbName, collectionName string) {
	d.databaseName = dbName
	d.collectionName = collectionName
}

// RequestRefreshAfterDatabaseDeletion is called by the modal component to signal
// to the dbcoltable component that a database deletion was successfully performed and
// that it should update its data and cursors accordingly
func (d *DbColTableState) RequestRefreshAfterDatabaseDeletion() {
	d.databaseDeletionSuccess = true
}

// RequestRefreshAfterCollectionDeletion is called by the modal component to signal
// to the dbcoltable component that a collection deletion was successfully performed and
// that it should update its data and cursor accordingly
func (d *DbColTableState) RequestRefreshAfterCollectionDeletion() {
	d.collectionDeletionSuccess = true
}

// WasDatabaseDeletedViaModal checks to see if the modal component successfully confirmed
// a deletion request from the user and then deleted a database. If a db was deleted,
// the dbcoltable component will then need to update its data
func (d *DbColTableState) WasDatabaseDeletedViaModal() bool {
	return d.databaseDeletionSuccess
}

// WasCollectionDeletedViaModal checks to see if the modal component successfully confirmed
// a deletion request from the user and then deleted a database. If a db was deleted,
// the dbcoltable component will then need to update its data
func (d *DbColTableState) WasCollectionDeletedViaModal() bool {
	return d.collectionDeletionSuccess
}

// ResetDatabaseDeletionRefreshFlag is called after the dbcoltable component has
// refreshed its data and cursors after a successful deletion to signal that
// it does not need to update its data again
func (d *DbColTableState) ResetDatabaseDeletionRefreshFlag() {
	d.databaseDeletionSuccess = false
}

// ResetCollectionDeletionRefreshFlag is called after the dbcoltable component has
// refreshed its data and cursors after a successful deletion to signal that
// it does not need to update its data again
func (d *DbColTableState) ResetCollectionDeletionRefreshFlag() {
	d.collectionDeletionSuccess = false
}
