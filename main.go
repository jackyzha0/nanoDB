package main

import (
	"github.com/jackyzha0/nanoDB/api"
	"github.com/jackyzha0/nanoDB/index"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Infof("initializing nanoDB")
	index.I.Regenerate("db")

    api.Serve()
}