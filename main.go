package main

import (
	"github.com/jacobmcgowan/svg-quick-edit/cmd"
	"github.com/spf13/afero"
)

func main() {
	cmd.Init(afero.NewOsFs())
	cmd.Execute()
}
