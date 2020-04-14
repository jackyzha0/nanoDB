// Package index ...
package index

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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
	// read lock
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
	// read lock
	i.mu.RLock()
	defer i.mu.RUnlock()

	// get if File exists, return nil and false otherwise
	if file, ok := i.index[key]; ok {
		return file, true
	}

	return &File{}, false
}

// ResolvePath returns a string representing the path to file
func (f *File) ResolvePath() string {
	return fmt.Sprintf("%s/%s.json", I.Dir, f.FileName)
}

// Regenerate rebuilds the current file index from given directory
// by crawling it for any .json files
func (i *FileIndex) Regenerate(dir string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	start := time.Now()
	log.Infof("building index for directory %s...", dir)

	i.Dir = dir
	i.index = i.buildIndexMap()
	log.Infof("built index in %d ms", time.Since(start).Milliseconds())
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
