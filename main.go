package main

import (
	"log"
	"os"

	"github.com/jacobmcgowan/svg-quick-edit/cmd"
	"github.com/spf13/afero"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLog := log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	fatalLog := log.New(os.Stderr, "FATAL: ", log.LstdFlags)

	cmd.Init(afero.NewOsFs(), infoLog, errorLog)
	if err := cmd.Execute(); err != nil {
		fatalLog.Printf("Error editing files: %s", err.Error())
		os.Exit(1)
	}
}
