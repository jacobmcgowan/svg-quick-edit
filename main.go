package main

import (
	"log"
	"os"

	"github.com/jacobmcgowan/svg-quick-edit/cmd"
	"github.com/spf13/afero"
)

func main() {
	cmd.Init(afero.NewOsFs())
	if err := cmd.Execute(); err != nil {
		log.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}
