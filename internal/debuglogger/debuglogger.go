// The debuglogger package is used during development to allow for the easy addition
// of slog.Info() logs to be printed to a mongotui.log file

package debuglogger

import (
	"fmt"
	"log/slog"
	"os"
)

func Initialize() *os.File {
	if _, ok := os.LookupEnv("MONGOTUI_DEBUG"); ok {
		file, err := os.Create("mongotui.log")
		if err != nil {
			fmt.Println("unable to create debug log file")
		}
		handler := slog.NewTextHandler(file, nil)
		logger := slog.New(handler)
		slog.SetDefault(logger)
		return file
	}
	return nil
}
