package modal

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ErrModalMsg:
		m.errMsg = &msg
	case DbCollInsertModalMsg:
		m.dbCollInsertMsg = &msg
		m.dbInsertInput.SetValue(msg.databaseName)
		m.collInsertInput.SetValue("")
		if msg.databaseName == "" {
			m.focusedDbCollInput = databaseInputFocused
			m.dbInsertInput.Focus()
			m.collInsertInput.Blur()
		} else {
			m.focusedDbCollInput = collectionInputFocused
			m.dbInsertInput.Blur()
			m.collInsertInput.Focus()
		}
	case CollDropModalMsg:
		m.collDropMsg = &msg
		m.confirmationCursor = yesButtonCursor
	case DbDropModalMsg:
		m.dbDropMsg = &msg
		m.confirmationCursor = yesButtonCursor
	case DocDeleteModalMsg:
		m.docDeleteMsg = &msg
		m.confirmationCursor = yesButtonCursor
	case DocInsertModalMsg:
		m.docInsertMsg = &msg
		m.confirmationCursor = yesButtonCursor
	case DocEditModalMsg:
		m.docEditMsg = &msg
		m.confirmationCursor = yesButtonCursor
	case tea.KeyMsg:
		m.errMsg = nil // Any key clears error messages
		if m.dbCollInsertMsg != nil {
			switch {
			case key.Matches(msg, keys.Enter):
				m.dbCollInsertMsg = nil
				return m, execDbCollectionInsert(m.dbInsertInput.Value(), m.collInsertInput.Value())
			case key.Matches(msg, keys.Tab):
				if m.focusedDbCollInput == databaseInputFocused {
					m.focusedDbCollInput = collectionInputFocused
					m.dbInsertInput.Blur()
					m.collInsertInput.Focus()
				} else {
					m.focusedDbCollInput = databaseInputFocused
					m.dbInsertInput.Focus()
					m.collInsertInput.Blur()
				}
				return m, nil
			case key.Matches(msg, keys.LineUp):
				m.focusedDbCollInput = databaseInputFocused
				m.dbInsertInput.Focus()
				m.collInsertInput.Blur()
				return m, nil
			case key.Matches(msg, keys.LineDown):
				m.focusedDbCollInput = collectionInputFocused
				m.dbInsertInput.Blur()
				m.collInsertInput.Focus()
				return m, nil
			default:
				if m.focusedDbCollInput == databaseInputFocused {
					m.dbInsertInput, _ = m.dbInsertInput.Update(msg)
				} else {
					m.collInsertInput, _ = m.collInsertInput.Update(msg)
				}
			}
		}

		switch {
		case key.Matches(msg, keys.Left):
			m.confirmationCursor = yesButtonCursor
		case key.Matches(msg, keys.Right):
			m.confirmationCursor = noButtonCursor
		case key.Matches(msg, keys.Enter):
			var cmd tea.Cmd
			if m.errMsg != nil {
				m.errMsg = nil
			} else if m.collDropMsg != nil {
				if m.confirmationCursor == yesButtonCursor {
					cmd = execCollectionDrop(m.collDropMsg.dbName, m.collDropMsg.collectionName)
				}
				m.collDropMsg = nil
				return m, cmd
			} else if m.dbDropMsg != nil {
				if m.confirmationCursor == yesButtonCursor {
					cmd = execDatabaseDrop(m.dbDropMsg.dbName)
				}
				m.dbDropMsg = nil
				return m, cmd
			} else if m.docDeleteMsg != nil {
				if m.confirmationCursor == yesButtonCursor {
					cmd = execDocDelete(m.docDeleteMsg.doc)
				}
				m.docDeleteMsg = nil
				return m, cmd
			} else if m.docInsertMsg != nil {
				if m.confirmationCursor == yesButtonCursor {
					cmd = execDocInsert(m.docInsertMsg.doc)
				}
				m.docInsertMsg = nil
				return m, cmd
			} else if m.docEditMsg != nil {
				if m.confirmationCursor == yesButtonCursor {
					cmd = execDocEdit(m.docEditMsg.oldDoc, m.docEditMsg.newDoc)
				}
				m.docEditMsg = nil
				return m, cmd
			}
		}
	}
	return m, nil
}
