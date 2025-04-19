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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const FILE_EXT = ".svg"

var fs afero.Fs
var finds []string
var replaces []string
var values []string
var suffixes []string
var path string
var exclude string
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "svg-quick-edit",
	Short: "Edits attributes of paths in SVG files.",
	Long: `svg-quick-edit is a CLI tool that allows you to quickly edit
attributes of paths in SVG files. It is useful for batch processing SVG files
to change attributes like 'fill', 'stroke', etc. The tool takes a path to an SVG
file or directory containing SVG files, and modifies the specified attributes
for all paths in the SVG file(s).

Example usage:
svg-quick-edit -f "class='aac-skin-fill'" -r fill -v "#e3ab72" -s new -p "/path/to/svg/files"

Will replace the 'fill' attribute of all paths with the class 'aac-skin-fill' in
the specified SVG files with the new value '#e3ab72' and saved the modified
images to new files with a suffix of "_new.svg". Images that are not modified
will not have a new file created.

You can also specify multiple find, replace, and value arguments to replace
attributes for multiple different paths in the SVG files. The number of find,
replace, and value arguments must be the same and must be in the same order.
For example:

svg-quick-edit -f "class='aac-skin-fill'" -r fill -v "#e3ab72" -s "skin-e3ab72" -f "class='aac-hair-fill'" -r fill -v "#a65e26" -s "hair-a65e26" -p "/path/to/svg/files"

Will replace the 'fill' attribute of all paths with the class
'aac-skin-fill' with the new value '#e3ab72' and the 'fill' attribute of all
paths with the class 'aac-hair-fill' with the new value '#a65e26'. The modified
images will be saved to new files with the suffixes
"_skin-e3ab72_hair-a65e26.svg" if both paths are found in the SVG file or
"_skin-e3ab72.svg" if only the first path is found and "_hair-a65e26.svg" if
only the second path is found.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(finds) != len(replaces) || len(finds) != len(values) || len(finds) != len(suffixes) {
			return fmt.Errorf("the number of find, replace, value, and suffix arguments must be the same")
		}

		if isFile := strings.HasSuffix(path, FILE_EXT); isFile {
			if err := editFile(path); err != nil {
				return fmt.Errorf("failed to edit SVG file: %s", err.Error())
			}
		} else {
			entries, err := afero.ReadDir(fs, path)
			if err != nil {
				return fmt.Errorf("failed to read directory: %s", err.Error())
			}

			for _, fileInfo := range entries {
				if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), FILE_EXT) {
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

func Init(appFs afero.Fs) {
	fs = appFs
}

func Execute() {
	if fs == nil {
		log.Print("File system not initialized. Please call Init() before Execute().")
		os.Exit(1)
	}

	err := rootCmd.Execute()
	if err != nil {
		log.Printf("Failed to edit SVG file(s): %s\n", err.Error())
		os.Exit(1)
	}
}

func GenMarkdownTree(path string) error {
	return doc.GenMarkdownTree(rootCmd, path)
}

func init() {
	rootCmd.Flags().StringSliceVarP(&finds, "find", "f", []string{}, "The attribute to find paths with.")
	rootCmd.Flags().StringSliceVarP(&replaces, "replace", "r", []string{}, "The attribute to replace the value for.")
	rootCmd.Flags().StringSliceVarP(&values, "value", "v", []string{}, "The new value to set the attribute to.")
	rootCmd.Flags().StringSliceVarP(&suffixes, "suffix", "s", []string{}, "The suffix to add to the modified SVG file name.")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "The path to the SVG file(s).")
	rootCmd.Flags().StringVarP(&exclude, "exclude", "e", "", "The regex of files to exclude.")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output.")

	rootCmd.MarkFlagRequired("find")
	rootCmd.MarkFlagRequired("replace")
	rootCmd.MarkFlagRequired("value")
	rootCmd.MarkFlagRequired("suffix")
	rootCmd.MarkFlagRequired("path")
	rootCmd.MarkFlagsRequiredTogether("find", "replace", "value", "suffix")
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

	file, err := fs.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open SVG file: %s", err.Error())
	}
	defer file.Close()

	doc, err := xmlquery.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse SVG file: %s", err.Error())
	}

	edited := make([]bool, len(suffixes))
	for i, find := range finds {
		paths := xmlquery.Find(doc, "//path[@"+find+"]")
		for _, path := range paths {
			path.SetAttr(replaces[i], values[i])
			edited[i] = true
		}
	}

	suffix := ""
	for i, didEdit := range edited {
		if didEdit {
			suffix = suffix + "_" + suffixes[i]
		}
	}

	if suffix == "" {
		if verbose {
			fmt.Printf("No modifications were made in %s\n", filepath)
		}

		return nil
	}

	i := strings.LastIndex(filepath, FILE_EXT)
	newFilepath := filepath[:i] + suffix + filepath[i:]
	if err = afero.WriteFile(fs, newFilepath, []byte(doc.OutputXML(true)), 0644); err != nil {
		return fmt.Errorf("failed to write modified SVG file %s: %s", newFilepath, err.Error())
	}

	if verbose {
		fmt.Printf("Modified SVG file saved as %s\n", newFilepath)
	}

	return nil
}
