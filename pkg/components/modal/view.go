package modal

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	// Error Modal
	if m.errMsg != nil {
		title := m.styles.ErrorHeader.Render("Error")
		return m.styles.Modal.Render(title + "\n\n" + m.errMsg.Err.Error())
	} else if m.dbCollInsertMsg != nil { // Database and collection insert modal
		text := "Enter the database and collection names you would like to insert\n"
		msg := fmt.Sprintf("%s\n\n%s\n%s", m.styles.InputTextBoxMsg.Render(text), m.styles.InputTextBox.Render(m.dbInsertInput.View()), m.styles.InputTextBox.Render(m.collInsertInput.View()))
		return m.styles.Modal.UnsetAlignHorizontal().Render(msg)
	} else { // All Confirmation modals
		var yesButton string
		var noButton string
		if m.confirmationCursor == yesButtonCursor {
			yesButton = m.styles.HighlightedButton.Render("Yes")
			noButton = m.styles.Button.Render("No")
		} else {
			yesButton = m.styles.Button.Render("Yes")
			noButton = m.styles.HighlightedButton.Render("No")
		}
		buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesButton, noButton)

		if m.collDropMsg != nil {
			title := m.styles.ConfirmationHeader.Render("Confirm")
			msg := fmt.Sprintf("%s\n\nAre you sure you would like to drop the collection %s?\n%s", title, m.collDropMsg.collectionName, buttons)
			return m.styles.Modal.Render(msg)
		} else if m.dbDropMsg != nil {
			title := m.styles.ConfirmationHeader.Render("Confirm")
			msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to drop the database %s?\n%s", title, m.dbDropMsg.dbName, buttons)
			return m.styles.Modal.Render(msg)
		} else if m.docDeleteMsg != nil {
			title := m.styles.ConfirmationHeader.Render("Confirm")
			msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to delete the selected document?\n%s", title, buttons)
			return m.styles.Modal.Render(msg)
		} else if m.docInsertMsg != nil {
			title := m.styles.ConfirmationHeader.Render("Confirm")
			msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to insert the new document?\n%s", title, buttons)
			return m.styles.Modal.Render(msg)
		} else if m.docEditMsg != nil {
			title := m.styles.ConfirmationHeader.Render("Confirm")
			msg := fmt.Sprintf("%s\n\n"+"Are you sure you would like to make your edits?\n%s", title, buttons)
			return m.styles.Modal.Render(msg)
		}
	}
	return ""
}
