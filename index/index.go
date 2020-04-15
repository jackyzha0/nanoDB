// Package index ...
package index

import (
	"fmt"
	"sync"
	"time"

	"github.com/jackyzha0/nanoDB/log"
)

// I is the global database index which keeps track of
// which files are where
var I *FileIndex

func init() {
	I = &FileIndex{
		Dir:   "",
		index: map[string]*File{},
	}
}

// FileIndex is holds the actual index mapping for keys to files
type FileIndex struct {
	mu    sync.RWMutex
	Dir   string
	index map[string]*File
}

// File stores the filename as well as a read-write mutex
type File struct {
	FileName string
	mu       sync.RWMutex
}

// List returns all keys in database
func (i *FileIndex) List() (res []string) {
	// read lock on index
	i.mu.RLock()
	defer i.mu.RUnlock()

	for k := range i.index {
		res = append(res, k)
	}

	return res
}

// Lookup returns the file with that key
// Returns (File, true) if file exists
// otherwise, returns new File, false
func (i *FileIndex) Lookup(key string) (*File, bool) {
	// read lock on index
	i.mu.RLock()
	defer i.mu.RUnlock()

	// get if File exists, return nil and false otherwise
	if file, ok := i.index[key]; ok {
		return file, true
	}

	return &File{FileName: key}, false
}

func (i *FileIndex) Put(file *File, bytes []byte) error {
	// write lock on index
	i.mu.Lock()
	defer i.mu.Unlock()

	i.index[file.FileName] = file
	err := file.ReplaceContent(string(bytes))
    return err
}

// ResolvePath returns a string representing the path to file
func (f *File) ResolvePath() string {
	if I.Dir == "" {
		return fmt.Sprintf("%s.json", f.FileName)
	}
	return fmt.Sprintf("%s/%s.json", I.Dir, f.FileName)
}

// Regenerate rebuilds the current file index from given directory
// by crawling it for any .json files
func (i *FileIndex) Regenerate(dir string) {
	// write lock on index
	i.mu.Lock()
	defer i.mu.Unlock()

	start := time.Now()
	log.Info("building index for directory %s...", dir)

	i.Dir = dir
	i.index = i.buildIndexMap()
	log.Info("built index of %d files in %d ms", len(i.index), time.Since(start).Milliseconds())
}

// creates a map from key to File
func (i *FileIndex) buildIndexMap() map[string]*File {
	newIndexMap := make(map[string]*File)

	files := crawlDirectory(i.Dir)
	for _, f := range files {
		newIndexMap[f] = &File{FileName: f}
	}

	return newIndexMap
}

// Delete deletes the given file and then removes it from I
func (i *FileIndex) Delete(file *File) error {
	// write lock on index
	i.mu.Lock()
	defer i.mu.Unlock()

	// delete first so pointer isn't nil
	err := file.Delete()

	if err == nil {
		delete(i.index, file.FileName)
	}

	return err
}
