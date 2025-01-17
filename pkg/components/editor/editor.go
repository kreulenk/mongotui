// The editor package contains code to edit a mongo document and then push it back to the database. It is not a traditional
// charmbracelet/bubbletea component as it needs to override stdout/stderr

package editor

import (
	"fmt"
	"github.com/kreulenk/mongotui/pkg/mongodata"
	"os"
	"os/exec"
	"path/filepath"
)

type Editor struct {
	engine *mongodata.Engine
}

func New(engine *mongodata.Engine) Editor {
	return Editor{
		engine: engine,
	}
}

func (e Editor) OpenFileInEditor() error {
	doc, err := e.engine.GetSelectedDocument()
	if err != nil {
		return err
	}
	file := filepath.Join(os.TempDir(), "mongoEdit.json")
	if err = os.WriteFile(file, doc, 0600); err != nil {
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
	return nil
}
