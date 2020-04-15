package index

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func crawlDirectory(directory string) []string {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	res := []string{}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".json" {
			name := strings.TrimSuffix(file.Name(), ".json")
			res = append(res, name)
		}
	}

	return res
}

// GetByteArray returns the byte array of given file
func (f *File) GetByteArray() ([]byte, error) {
	return ioutil.ReadFile(f.ResolvePath())
}

// changes the contents of file f to be str
func (f *File) replaceContent(str string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// create blank file
	_, err := os.Create(f.ResolvePath())
	if err != nil {
		return err
	}

	file, err := os.OpenFile(f.ResolvePath(), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	defer file.Close()

	// appends the given str to the now empty file
	_, err = file.WriteString(str)
	if err != nil {
		return err
	}

	// success
	return nil
}
