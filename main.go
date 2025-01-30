package main

import (
	"github.com/kreulenk/mongotui/cmd"
	"github.com/kreulenk/mongotui/internal/debuglogger"
)

func main() {
	logFile := debuglogger.Initialize()
	defer logFile.Close()
	cmd.Execute()
}
