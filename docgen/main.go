package main

import (
	"log"

	"github.com/jacobmcgowan/svg-quick-edit/cmd"
)

func main() {
	if err := cmd.GenMarkdownTree("../docs"); err != nil {
		log.Fatalf("Failed to generate documentation: %s", err)
	}
}
