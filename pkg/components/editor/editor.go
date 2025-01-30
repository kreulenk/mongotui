// The editor package contains code to edit a mongo document and then push it back to the database. It is not a traditional
// charmbracelet/bubbletea component as it needs to override stdout/stderr

package editor

import (
	"fmt"
	"github.com/kreulenk/mongotui/internal/state"
	"github.com/kreulenk/mongotui/pkg/mongoengine"
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

func (e Editor) OpenFileInEditor() error {
	oldDoc, err := e.engine.GetSelectedDocumentMarshalled()
	if err != nil {
		return err
	}
	file := filepath.Join(os.TempDir(), "mongoEdit.json")
	if err = os.WriteFile(file, oldDoc, 0600); err != nil {
		return fmt.Errorf("failed to write file to allow for doc editing: %w", err)
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
	if err = openEditorCmd.Run(); err != nil {
		return fmt.Errorf("error opening %s editor: %v", editor, err)
	}

	newDoc, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read the file that was just edited: %w", err)
	}

	if err = e.engine.UpdateDocument(oldDoc, newDoc); err != nil {
		return err
	}

	e.state.SetActiveComponent(state.DocList)
	return nil
}
