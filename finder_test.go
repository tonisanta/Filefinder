package filefinder

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

const (
	greeting                  = "greeting.txt"
	greetingInSubdirectory    = "subdirectory/greeting2.txt"
	greetingInSubSubdirectory = "subdirectory/another/greeting3.txt"
)

var testFS = fstest.MapFS{
	greeting:                  {Data: []byte("Hello world")},
	"fruit.txt":               {Data: []byte("apple, orange")},
	greetingInSubdirectory:    {Data: []byte("Hello gophers")},
	greetingInSubSubdirectory: {Data: []byte("This file also contains Hello")},
}

func TestFinder(t *testing.T) {

	t.Run("return the file paths of the files containing the word Hello", func(t *testing.T) {
		outChan, err := FindFiles(testFS, "Hello")
		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}

		var got []string
		for file := range outChan {
			got = append(got, file)
		}

		expected := []string{greeting, greetingInSubdirectory, greetingInSubSubdirectory}
		for _, file := range expected {
			assertContainsPath(t, got, file)
		}
	})

	t.Run("return error if directory doesn't exist", func(t *testing.T) {
		root := "/non-existing/directory"
		fileSystem := os.DirFS(root)
		_, err := FindFiles(fileSystem, "Hello")
		if err == nil {
			t.Errorf("The function should return an error")
		}
	})

}

func assertContainsPath(t testing.TB, paths []string, path string) {
	t.Helper()
	contains := false
	for _, x := range paths {
		// as the test FS uses slash '/', we have to compare with the os Separator
		if x == filepath.FromSlash(path) {
			contains = true
			return
		}
	}
	if !contains {
		t.Errorf("expected %s to contain %q but it didn't", paths, path)
	}
}

func ExampleFindFiles() {
	outChan, err := FindFiles(testFS, "Hello")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	for file := range outChan {
		fmt.Println(file)
	}

	// Output: greeting.txt
	// subdirectory\greeting2.txt
	// subdirectory\another\greeting3.txt
}

/*
	------- BENCHMARK -------
	We are going to compare solving the same problem using the FindFiles function
	compared with the one that provides the standard library to traverse a directory
	(WalkDir). To do so we are generating a test FileSystem with a custom number of
	files and words.
*/

const (
	numFilesBenchFS = 3_000
	numWordsPerFile = 10_000
	wordsPerLine    = 30
)

func BenchmarkWalkDirStandardLibrary(b *testing.B) {
	benchFS := generateBenchFS(numFilesBenchFS, numWordsPerFile)
	for i := 0; i < b.N; i++ {
		FindFilesStandardLibrary(benchFS, ".", "customWord")
	}
}

func BenchmarkWalkDir(b *testing.B) {
	benchFS := generateBenchFS(numFilesBenchFS, numWordsPerFile)
	for i := 0; i < b.N; i++ {
		outChan, _ := FindFiles(benchFS, "customWord")
		for range outChan {
		}
	}
}

func generateBenchFS(numFiles, numWordsPerFile int) fstest.MapFS {
	benchFS := make(fstest.MapFS, numFiles)
	fileName := ""
	for i := 0; i < numFiles; i++ {
		fileName = fmt.Sprintf("file%d", i)
		var data []byte
		for i := 0; i < numWordsPerFile; i++ {
			data = append(data, []byte("something ")...)
			if (i+1)%wordsPerLine == 0 {
				data = append(data, []byte("\n")...)
			}
		}

		benchFS[fileName] = &fstest.MapFile{Data: data}
	}
	return benchFS
}

func FindFilesStandardLibrary(fileSystem fs.FS, directory string, word string) []string {
	var files []string
	fs.WalkDir(fileSystem, directory, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			fileName := filepath.Join(path, entry.Name())
			containsWord, err := fileContainsWord(fileSystem, path, word)
			if err != nil {
				return err
			}
			if containsWord {
				files = append(files, fileName)
			}
		}
		return nil
	})

	return files
}

func fileContainsWord(fileSystem fs.FS, fileName string, word string) (bool, error) {
	file, err := fileSystem.Open(fileName)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(line, word) {
			return true, nil
		}
	}
	return false, nil
}
