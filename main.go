package main

import (
	"log"
	"os"

	"github.com/jacobmcgowan/svg-quick-edit/cmd"
	"github.com/spf13/afero"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO: ", 0)
	errorLog := log.New(os.Stderr, "ERROR: ", 0)

	cmd.Init(afero.NewOsFs(), infoLog, errorLog)
	if err := cmd.Execute(); err != nil {
		errorLog.Printf("Error editing files: %s", err.Error())
		os.Exit(1)
	}
}
