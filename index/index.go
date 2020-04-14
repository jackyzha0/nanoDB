package index

import (
	"sync"
)

var I *FileIndex

func init() {
	I = &FileIndex{
		dir: "",
		index: map[string]File{},
	}
}

// FileIndex
type FileIndex struct {
	mu    sync.RWMutex
	dir   string
	index map[string]File
}

type File struct {
	FileName string
	mu       sync.RWMutex
}

func (i FileIndex) Lookup(key string) (File, bool) {
	// read lock
	i.mu.RLock()
	defer i.mu.RUnlock()

	// get if File exists, return nil and false otherwise
	if file, ok := i.index[key]; ok {
		return file, true
	}

	return File{}, false
}

func (i FileIndex) Regenerate(dir string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.dir = dir
	i.index = i.buildIndexMap()
}

func (i FileIndex) buildIndexMap() map[string]File {
	newIndexMap := make(map[string]File)

	files := crawlDirectory(i.dir)
	for _, f := range files {
		newIndexMap[f] = File{FileName: f}
	}

	return newIndexMap
}
