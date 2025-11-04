package cmd

import (
	"log"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type fakeWriter struct {
	Logs []string
}

func (fw *fakeWriter) Write(p []byte) (n int, err error) {
	fw.Logs = append(fw.Logs, string(p))
	return len(p), nil
}

func TestExecute(t *testing.T) {
	fs := afero.NewMemMapFs()
	fakeStdout := fakeWriter{}
	fakeStderr := fakeWriter{}
	fakeInfoLog := log.New(&fakeStdout, "", 0)
	fakeErrorLog := log.New(&fakeStderr, "", 0)

	Init(fs, fakeInfoLog, fakeErrorLog)

	cleanLogs := func() {
		fakeStdout.Logs = []string{}
		fakeStderr.Logs = []string{}
	}

	createMockFile := func(path, content string) {
		err := afero.WriteFile(fs, path, []byte(content), 0644)
		require.NoError(t, err)
	}

	assertExists := func(path string) {
		exists, err := afero.Exists(fs, path)
		require.NoError(t, err)
		require.True(t, exists, "File should exist: "+path)
	}

	assertNotExists := func(path string) {
		exists, err := afero.Exists(fs, path)
		require.NoError(t, err)
		require.False(t, exists, "File should not exist: "+path)
	}

	assertContains := func(path string, content []string) {
		fileContent, err := afero.ReadFile(fs, path)
		require.NoError(t, err)
		for _, c := range content {
			require.Contains(t, string(fileContent), c, "File content should contain: "+c)
		}
	}

	assertNotContains := func(path string, content []string) {
		fileContent, err := afero.ReadFile(fs, path)
		require.NoError(t, err)
		for _, c := range content {
			require.NotContains(t, string(fileContent), c, "File content should not contain: "+c)
		}
	}

	t.Run("SingleSVGFile", func(t *testing.T) {
		cleanLogs()
		filePath := "/test.svg"
		createMockFile(filePath, `<svg><path class="test-class" fill="#ffffff" /></svg>`)

		finds = []string{"class='test-class'"}
		replaces = []string{"fill"}
		values = []string{"#000000"}
		suffixes = []string{"modified"}
		path = filePath
		exclude = ""

		err := execute()
		require.NoError(t, err)

		modifiedFilePath := "/test_modified.svg"
		assertExists(modifiedFilePath)
		assertContains(modifiedFilePath, []string{`fill="#000000"`})
	})

	t.Run("DirectoryWithSVGFiles", func(t *testing.T) {
		cleanLogs()
		dirPath := "/svg-files"
		fs.Mkdir(dirPath, 0755)
		createMockFile(dirPath+"/file1.svg", `<svg><path class="test-class" fill="#ffffff" /></svg>`)
		createMockFile(dirPath+"/file2.svg", `<svg><path class="test-class" fill="#ffffff" /></svg>`)

		finds = []string{"class='test-class'"}
		replaces = []string{"fill"}
		values = []string{"#123456"}
		suffixes = []string{"updated"}
		path = dirPath
		exclude = ""

		err := execute()
		require.NoError(t, err)

		modifiedFile1 := dirPath + "/file1_updated.svg"
		assertExists(modifiedFile1)
		assertContains(modifiedFile1, []string{`fill="#123456"`})

		modifiedFile2 := dirPath + "/file2_updated.svg"
		assertExists(modifiedFile2)
		assertContains(modifiedFile2, []string{`fill="#123456"`})
	})

	t.Run("ExcludeFiles", func(t *testing.T) {
		cleanLogs()
		dirPath := "/exclude-test"
		fs.Mkdir(dirPath, 0755)
		createMockFile(dirPath+"/include.svg", `<svg><path class="test-class" fill="#ffffff" /></svg>`)
		createMockFile(dirPath+"/exclude.svg", `<svg><path class="test-class" fill="#ffffff" /></svg>`)

		finds = []string{"class='test-class'"}
		replaces = []string{"fill"}
		values = []string{"#654321"}
		suffixes = []string{"modified"}
		path = dirPath
		exclude = "exclude\\.svg"

		err := execute()
		require.NoError(t, err)

		modifiedInclude := dirPath + "/include_modified.svg"
		assertExists(modifiedInclude)

		modifiedExclude := dirPath + "/exclude_modified.svg"
		assertNotExists(modifiedExclude)
	})

	t.Run("MultipleFindAndReplaces", func(t *testing.T) {
		cleanLogs()
		dirPath := "/multiple-replace-test"
		fs.Mkdir(dirPath, 0755)
		createMockFile(dirPath+"/file1.svg", `<svg><path class="a" fill="#ffffff" /></svg>`)
		createMockFile(dirPath+"/file2.svg", `<svg><path class="b" stroke="#ffffff" /></svg>`)
		createMockFile(dirPath+"/file3.svg", `<svg><path class="a" fill="#ffffff" /><path class="b" stroke="#ffffff" /></svg>`)
		createMockFile(dirPath+"/file4.svg", `<svg><path class="c" fill="#ffffff" /></svg>`)

		finds = []string{"class='a'", "class='b'"}
		replaces = []string{"fill", "stroke"}
		values = []string{"#111111", "#222222"}
		suffixes = []string{"a-modified", "b-modified"}
		path = dirPath
		exclude = ""

		err := execute()
		require.NoError(t, err)

		modifiedFile1s := []string{
			dirPath + "/file1_a-modified.svg",
			dirPath + "/file1_b-modified.svg",
			dirPath + "/file1_a-modified_b-modified.svg",
		}
		shouldExist1 := []bool{
			true,
			false,
			false,
		}
		shouldContain1 := []string{
			`fill="#111111"`,
		}
		shouldNotContain1 := []string{
			`stroke="#222222"`,
		}
		modifiedFile2s := []string{
			dirPath + "/file2_a-modified.svg",
			dirPath + "/file2_b-modified.svg",
			dirPath + "/file2_a-modified_b-modified.svg",
		}
		shouldExist2 := []bool{
			false,
			true,
			false,
		}
		shouldContain2 := []string{
			`stroke="#222222"`,
		}
		shouldNotContain2 := []string{
			`fill="#111111"`,
		}
		modifiedFile3s := []string{
			dirPath + "/file3_a-modified.svg",
			dirPath + "/file3_b-modified.svg",
			dirPath + "/file3_a-modified_b-modified.svg",
		}
		shouldExist3 := []bool{
			false,
			false,
			true,
		}
		shouldContain3 := []string{
			`fill="#111111"`,
			`stroke="#222222"`,
		}
		modifiedFile4s := []string{
			dirPath + "/file4_a-modified.svg",
			dirPath + "/file4_b-modified.svg",
			dirPath + "/file4_a-modified_b-modified.svg",
		}

		for i := range 3 {
			if shouldExist1[i] {
				assertExists(modifiedFile1s[i])
				assertContains(modifiedFile1s[i], shouldContain1)
				assertNotContains(modifiedFile1s[i], shouldNotContain1)

			} else {
				assertNotExists(modifiedFile1s[i])
			}

			if shouldExist2[i] {
				assertExists(modifiedFile2s[i])
				assertContains(modifiedFile2s[i], shouldContain2)
				assertNotContains(modifiedFile2s[i], shouldNotContain2)

			} else {
				assertNotExists(modifiedFile2s[i])
			}

			if shouldExist3[i] {
				assertExists(modifiedFile3s[i])
				assertContains(modifiedFile3s[i], shouldContain3)

			} else {
				assertNotExists(modifiedFile3s[i])
			}

			assertNotExists(modifiedFile4s[i])
		}
	})

	t.Run("FileNotFoundTest", func(t *testing.T) {
		cleanLogs()
		filePath := "/not-found.svg"

		finds = []string{"class='test-class'"}
		replaces = []string{"fill"}
		values = []string{"#000000"}
		suffixes = []string{"modified"}
		path = filePath
		exclude = ""

		err := execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "encountered 1 error while editing SVG files")

		require.Len(t, fakeStderr.Logs, 1)
		require.Contains(t, fakeStderr.Logs[0], "failed to open SVG file")
	})

	t.Run("DirectoryNotFoundTest", func(t *testing.T) {
		cleanLogs()
		filePath := "/not-found"

		finds = []string{"class='test-class'"}
		replaces = []string{"fill"}
		values = []string{"#000000"}
		suffixes = []string{"modified"}
		path = filePath
		exclude = ""

		err := execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "encountered 1 error while editing SVG files")

		require.Len(t, fakeStderr.Logs, 1)
		require.Contains(t, fakeStderr.Logs[0], "Failed to read directory")
	})

	t.Run("NonEqualArgumentsCount", func(t *testing.T) {
		cleanLogs()
		filePath := "/not-equal.svg"
		createMockFile(filePath, `<svg><path class="test-class" fill="#ffffff" /></svg>`)

		argsTooManyFinds := map[string][]string{
			"finds":    {"class='test-class'", "class='another-class'"},
			"replaces": {"fill"},
			"values":   {"#000000"},
			"suffixes": {"modified"},
		}
		argsTooManyReplaces := map[string][]string{
			"finds":    {"class='test-class'"},
			"replaces": {"fill", "stroke"},
			"values":   {"#000000"},
			"suffixes": {"modified"},
		}
		argsTooManyValues := map[string][]string{
			"finds":    {"class='test-class'"},
			"replaces": {"fill"},
			"values":   {"#000000", "#ffffff"},
			"suffixes": {"modified"},
		}
		argsTooManySuffixes := map[string][]string{
			"finds":    {"class='test-class'"},
			"replaces": {"fill"},
			"values":   {"#000000"},
			"suffixes": {"modified", "another-modified"},
		}

		path = filePath
		exclude = ""

		finds = argsTooManyFinds["finds"]
		replaces = argsTooManyFinds["replaces"]
		values = argsTooManyFinds["values"]
		suffixes = argsTooManyFinds["suffixes"]
		err := execute()
		require.Error(t, err)

		finds = argsTooManyReplaces["finds"]
		replaces = argsTooManyReplaces["replaces"]
		values = argsTooManyReplaces["values"]
		suffixes = argsTooManyReplaces["suffixes"]
		err = execute()
		require.Error(t, err)

		finds = argsTooManyValues["finds"]
		replaces = argsTooManyValues["replaces"]
		values = argsTooManyValues["values"]
		suffixes = argsTooManyValues["suffixes"]
		err = execute()
		require.Error(t, err)

		finds = argsTooManySuffixes["finds"]
		replaces = argsTooManySuffixes["replaces"]
		values = argsTooManySuffixes["values"]
		suffixes = argsTooManySuffixes["suffixes"]
		err = execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "the number of find, replace, value, and suffix arguments must be the same")
	})
}
