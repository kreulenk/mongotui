package modal

import "fmt"

func (m *Model) View() string {
	if m.errMsg != nil {
		title := m.styles.ErrorHeader.Render("Error")
		return m.styles.Modal.Render(title + "\n\n" + m.errMsg.Err.Error())
	} else if m.collDropMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\nAre you sure you would like to drop the collection %s?\nPress Enter to confirm.", title, m.collDropMsg.collectionName)
		return m.styles.Modal.Render(msg)
	} else if m.dbDropMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to drop the database %s?\nPress Enter to confirm.", title, m.dbDropMsg.dbName)
		return m.styles.Modal.Render(msg)
	} else if m.docDeleteMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to delete the selected document?\nPress Enter to confirm.", title)
		return m.styles.Modal.Render(msg)
	} else if m.docInsertMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to insert the new document?\nPress Enter to confirm.", title)
		return m.styles.Modal.Render(msg)
	} else if m.docEditMsg != nil {
		title := m.styles.ConfirmationHeader.Render("Confirm")
		msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to make your edits?\nPress Enter to confirm.", title)
		return m.styles.Modal.Render(msg)
	}
	return ""
}
