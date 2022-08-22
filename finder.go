package filefinder

import (
	"bufio"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// FindFiles returns the relative paths of the files containing the word you are looking for
func FindFiles(fileSystem fs.FS, word string) (<-chan string, error) {
	outChan := make(chan string)
	wg := sync.WaitGroup{}
	err := checkDirectory(fileSystem, ".", word, &wg, outChan)
	if err != nil {
		return nil, err
	}

	go func() {
		// when all the goroutines finish, the channel is closed
		wg.Wait()
		close(outChan)
	}()

	return outChan, nil
}

func checkDirectory(fileSystem fs.FS, directory, word string, wg *sync.WaitGroup, outChan chan string) error {
	dirEntries, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return err
	}

	for _, entry := range dirEntries {
		wg.Add(1)
		go func(entry fs.DirEntry) {
			defer wg.Done()
			if entry.IsDir() {
				dir := filepath.Join(directory, entry.Name())
				subFS, err := fs.Sub(fileSystem, entry.Name())
				if err != nil {
					return
				}
				_ = checkDirectory(subFS, dir, word, wg, outChan)
			} else {
				checkFile(fileSystem, directory, entry.Name(), word, outChan)
			}
		}(entry)
	}
	return nil
}

func checkFile(fileSystem fs.FS, directory, fileName, word string, out chan string) {
	file, err := fileSystem.Open(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(line, word) {
			out <- filepath.Join(directory, fileName)
		}
	}
}
