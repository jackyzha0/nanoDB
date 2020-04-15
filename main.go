package main

import (
	"github.com/jackyzha0/nanoDB/api"
	"github.com/jackyzha0/nanoDB/index"

	"github.com/jackyzha0/nanoDB/log"
)

func main() {
	log.SetLoggingLevel(log.INFO)
	log.Info("initializing nanoDB")
	index.I = index.NewFileIndex("db")
	index.I.Regenerate()

    api.Serve()
}