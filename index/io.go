package index

import (
	"encoding/json"
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

func (f *File) ToMap() (res map[string]interface{}, err error) {
	// get bytes
	bytes, err := f.getByteArray()
	if err != nil {
		return res, err
	}

	// unmarshal into map
	err = json.Unmarshal(bytes, &res)
	return res, err
}

// GetByteArray returns the byte array of given file
func (f *File) getByteArray() ([]byte, error) {
	// read lock on file
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	return ioutil.ReadFile(f.ResolvePath())
}

// ReplaceContent changes the contents of file f to be str
func (f *File) ReplaceContent(str string) error {
	// write lock on file
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

// Delete tries to remove the file
func (f *File) Delete() error {
	// write lock on file
	f.mu.Lock()
	defer f.mu.Unlock()

	// tries to delete the file
	err := os.Remove(f.ResolvePath())
	if err != nil {
		return err
	}

	return nil
}
