package modal

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ErrModalMsg:
		m.errMsg = &msg
	case CollDropModalMsg:
		m.collDropMsg = &msg
	case DbDropModalMsg:
		m.dbDropMsg = &msg
	case DocDeleteModalMsg:
		m.docDeleteMsg = &msg
	case DocInsertModalMsg:
		m.docInsertMsg = &msg
	case DocEditModalMsg:
		m.docEditMsg = &msg
	case tea.KeyMsg:
		if !m.IsModalDisplaying() {
			return m, nil
		}
		switch msg.Type {
		case tea.KeyEnter:
			if m.errMsg != nil {
				m.errMsg = nil
			} else if m.collDropMsg != nil {
				execCmd := execCollectionDrop(m.collDropMsg.dbName, m.collDropMsg.collectionName)
				m.collDropMsg = nil
				return m, execCmd
			} else if m.dbDropMsg != nil {
				execCmd := execDatabaseDrop(m.dbDropMsg.dbName)
				m.dbDropMsg = nil
				return m, execCmd
			} else if m.docDeleteMsg != nil {
				execCmd := execDocDelete(m.docDeleteMsg.doc)
				m.docDeleteMsg = nil
				return m, execCmd
			} else if m.docInsertMsg != nil {
				execCmd := execDocInsert(m.docInsertMsg.doc)
				m.docInsertMsg = nil
				return m, execCmd
			} else if m.docEditMsg != nil {
				execCmd := execDocEdit(m.docEditMsg.oldDoc, m.docEditMsg.newDoc)
				m.docEditMsg = nil
				return m, execCmd
			}
		default: // If it is a confirmation modal and enter was not selected, exit the modal with no actions performed
			m.errMsg = nil
			m.collDropMsg = nil
			m.dbDropMsg = nil
			m.docDeleteMsg = nil
			m.docInsertMsg = nil
			m.docEditMsg = nil
		}
	}
	return m, nil
}
