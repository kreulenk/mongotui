// The editor package contains code to insert/edit a mongo document and then push it back to the database.
// It is not a traditional charmbracelet/bubbletea component as it needs to override stdout/stderr

package editor

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kreulenk/mongotui/pkg/components/modal"
	"github.com/kreulenk/mongotui/pkg/mainview/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
	"go.mongodb.org/mongo-driver/v2/bson"
	"os"
	"os/exec"
	"path/filepath"
)

type Editor struct {
	state  *state.MainViewState
	engine *mongoengine.Engine
}

func New(engine *mongoengine.Engine, state *state.MainViewState) Editor {
	return Editor{
		state:  state,
		engine: engine,
	}
}

func (e Editor) EditDoc() tea.Cmd {
	oldDoc, err := e.engine.GetSelectedDocumentMarshalled()
	if err != nil {
		return modal.DisplayErrorModal(err)
	}
	newDoc, err := e.openFileInEditor(oldDoc)
	if err != nil {
		return modal.DisplayErrorModal(err)
	}

	var oldDocBson bson.M
	if err := bson.UnmarshalExtJSON(oldDoc, false, &oldDocBson); err != nil {
		return modal.DisplayErrorModal(fmt.Errorf("failed to parse the original document needed for the replacement: %w", err))
	}
	var newDocBson bson.M
	if err := bson.UnmarshalExtJSON(newDoc, false, &newDocBson); err != nil {
		return modal.DisplayErrorModal(fmt.Errorf("failed to parse the new document needed for the replacement: %w", err))
	}

	return modal.DisplayDocEditModal(oldDocBson, newDocBson)
}

func (e Editor) InsertDoc() tea.Cmd {
	newDoc := make(bson.M)
	newDoc["_id"] = bson.NewObjectID()
	newDocBytes, err := bson.MarshalExtJSONIndent(newDoc, false, false, "", "  ")
	if err != nil {
		return modal.DisplayErrorModal(fmt.Errorf("failed to marshal new document: %w", err))
	}

	editedDocBytes, err := e.openFileInEditor(newDocBytes)
	if err != nil {
		return modal.DisplayErrorModal(err)
	}

	var editedDoc bson.M
	if err := bson.UnmarshalExtJSON(editedDocBytes, false, &editedDoc); err != nil {
		return modal.DisplayErrorModal(fmt.Errorf("failed to unmarshal edited document: %w", err))
	}

	return modal.DisplayDocInsertModal(editedDoc)
}

func (e Editor) openFileInEditor(doc []byte) ([]byte, error) {
	file := filepath.Join(os.TempDir(), "mongoEdit.json")
	if err := os.WriteFile(file, doc, 0600); err != nil {
		return nil, fmt.Errorf("failed to write file to allow for doc editing: %w", err)
	}
	defer os.Remove(file)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	openEditorCmd := exec.Command(editor, file)
	openEditorCmd.Stdout = os.Stdout
	openEditorCmd.Stdin = os.Stdin
	openEditorCmd.Stderr = os.Stderr
	if err := openEditorCmd.Run(); err != nil {
		return nil, fmt.Errorf("error opening %s editor: %v", editor, err)
	}

	newDoc, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file that was just edited: %w", err)
	}

	return newDoc, nil
}
