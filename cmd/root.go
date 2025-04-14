/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var find string
var replace string
var value string
var path string
var suffix string

var rootCmd = &cobra.Command{
	Use:   "svg-quick-edit",
	Short: "Edits an attribute of paths in SVG files.",
	Long: `svg-quick-edit is a command line tool that allows you to quickly edit
an attribute of paths in SVG files. It is useful for batch processing SVG files
to change attributes like 'fill', 'stroke', etc. The tool takes a path to an SVG
file or directory containing SVG files, and modifies the specified attribute
for all paths in the SVG files.

Example usage:
svg-quick-edit -f "class='aac-skin-fill'" -r fill -v "#e3ab72" -p "/path/to/svg/files"

Output:
The modified SVG file(s) will be saved with a new name, which is the original
file name with a suffix added. For example, if the original filename is
"icon.svg", the modified file will be saved as "icon_new.svg". The suffix can be
specified using the -s flag. If not specified, the default suffix is "new".
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("find %s", find)

		if isFile := strings.HasSuffix(path, ".svg"); isFile {
			if err := editFile(path); err != nil {
				return fmt.Errorf("failed to edit SVG file: %s", err.Error())
			}
		} else {
			entries, err := os.ReadDir(path)
			if err != nil {
				return fmt.Errorf("failed to read directory: %s", err.Error())
			}

			for _, entry := range entries {
				fileInfo, err := entry.Info()
				if err != nil {
					return fmt.Errorf("failed to get file info: %s", err.Error())
				}

				if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".svg") {
					filepath := path + "/" + fileInfo.Name()
					if err := editFile(filepath); err != nil {
						return fmt.Errorf("failed to edit SVG file %s: %s", filepath, err.Error())
					}
				}
			}
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Printf("Failed to edit SVG file(s): %s", err.Error())
		os.Exit(1)
	}
}

func GenMarkdownTree(path string) error {
	return doc.GenMarkdownTree(rootCmd, path)
}

func init() {
	rootCmd.Flags().StringVarP(&find, "find", "f", "", "The attribute to find paths with.")
	rootCmd.MarkFlagRequired("find")
	rootCmd.Flags().StringVarP(&replace, "replace", "r", "", "The attribute to replace the value for.")
	rootCmd.MarkFlagRequired("replace")
	rootCmd.Flags().StringVarP(&value, "value", "v", "", "The new value to set the attribute to.")
	rootCmd.MarkFlagRequired("value")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "The path to the SVG file(s).")
	rootCmd.MarkFlagRequired("path")
	rootCmd.Flags().StringVarP(&suffix, "suffix", "s", "new", "The suffix to add to the modified SVG file name.")
}

func editFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open SVG file: %s", err.Error())
	}
	defer file.Close()

	doc, err := xmlquery.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse SVG file: %s", err.Error())
	}

	paths := xmlquery.Find(doc, "//path[@"+find+"]")
	for _, path := range paths {
		path.SetAttr(replace, value)
	}

	i := strings.LastIndex(filepath, ".svg")
	newFilepath := filepath[:i] + "_" + suffix + ".svg"
	if err = os.WriteFile(newFilepath, []byte(doc.OutputXML(true)), 0644); err != nil {
		return fmt.Errorf("failed to write modified SVG file %s: %s", newFilepath, err.Error())
	}

	return nil
}
