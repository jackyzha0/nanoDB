package main

import (
	_ "github.com/jackyzha0/nanoDB/api"
	"github.com/jackyzha0/nanoDB/index"
)

func main() {
	index.I.Regenerate()
}