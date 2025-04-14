/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var finds []string
var replaces []string
var values []string
var path string
var exclude string
var suffix string

var rootCmd = &cobra.Command{
	Use:   "svg-quick-edit",
	Short: "Edits attributes of paths in SVG files.",
	Long: `svg-quick-edit is a CLI tool that allows you to quickly edit
attributes of paths in SVG files. It is useful for batch processing SVG files
to change attributes like 'fill', 'stroke', etc. The tool takes a path to an SVG
file or directory containing SVG files, and modifies the specified attributes
for all paths in the SVG file(s).

Example usage:
svg-quick-edit -f "class='aac-skin-fill'" -r fill -v "#e3ab72" -p "/path/to/svg/files"

Will replace the 'fill' attribute of all paths with the class 'aac-skin-fill' in
the specified SVG files with the new value '#e3ab72'.

You can also specify multiple find, replace, and value arguments to replace
attributes for multiple different paths in the SVG files. The number of find,
replace, and value arguments must be the same and must be in the same order.
For example:

svg-quick-edit -f "class='aac-skin-fill'" -r fill -v "#e3ab72" -f "class='aac-hair-fill'" -r fill -v "#a65e26" -p "/path/to/svg/files"

Will replace the 'fill' attribute of all paths with the class
'aac-skin-fill' with the new value '#e3ab72' and the 'fill' attribute of all
paths with the class 'aac-hair-fill' with the new value '#a65e26'.

Output:
The modified SVG file(s) will be saved with a new name, which is the original
file name with a suffix added. For example, if the original filename is
"icon.svg", the modified file will be saved as "icon_new.svg". The suffix can be
specified using the -s flag. If not specified, the default suffix is "new".
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(finds) != len(replaces) || len(finds) != len(values) {
			return fmt.Errorf("the number of find, replace, and value arguments must be the same")
		}

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
	rootCmd.Flags().StringSliceVarP(&finds, "find", "f", []string{}, "The attribute to find paths with.")
	rootCmd.MarkFlagRequired("find")
	rootCmd.Flags().StringSliceVarP(&replaces, "replace", "r", []string{}, "The attribute to replace the value for.")
	rootCmd.MarkFlagRequired("replace")
	rootCmd.Flags().StringSliceVarP(&values, "value", "v", []string{}, "The new value to set the attribute to.")
	rootCmd.MarkFlagRequired("value")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "The path to the SVG file(s).")
	rootCmd.MarkFlagRequired("path")
	rootCmd.Flags().StringVarP(&exclude, "exclude", "e", "", "The regex of files to exclude.")
	rootCmd.Flags().StringVarP(&suffix, "suffix", "s", "new", "The suffix to add to the modified SVG file name.")
}

func editFile(filepath string) error {
	if exclude != "" {
		shouldExclude, err := regexp.MatchString(exclude, filepath)
		if err != nil {
			return fmt.Errorf("failed to compile regex %s: %s", exclude, err.Error())
		}
		if shouldExclude {
			return nil
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open SVG file: %s", err.Error())
	}
	defer file.Close()

	doc, err := xmlquery.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse SVG file: %s", err.Error())
	}

	for i, find := range finds {
		paths := xmlquery.Find(doc, "//path[@"+find+"]")
		for _, path := range paths {
			path.SetAttr(replaces[i], values[i])
		}
	}

	i := strings.LastIndex(filepath, ".svg")
	newFilepath := filepath[:i] + "_" + suffix + ".svg"
	if err = os.WriteFile(newFilepath, []byte(doc.OutputXML(true)), 0644); err != nil {
		return fmt.Errorf("failed to write modified SVG file %s: %s", newFilepath, err.Error())
	}

	return nil
}
