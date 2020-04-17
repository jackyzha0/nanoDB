package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/jackyzha0/nanoDB/log"
	"github.com/jackyzha0/nanoDB/index"
)

func shell(dir string) error {
	log.IsShellMode = true
	log.Info("starting nanodb shell...")
	setup(dir)

	reader := bufio.NewReader(os.Stdin)
	for {
		// input indicator
		log.Prompt("nanodb> ")

		// read keyboad input
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Warn("err reading input: %s", err.Error())
		}

		// Handle the execution of the input.
		if err = execInput(input); err != nil {
			log.Warn("err executing input: %s", err.Error())
		}
	}
}

func execInput(input string) (err error) {
	input = strings.TrimSuffix(input, "\n")
	args := strings.Split(input, " ")

	switch args[0] {
	case "index":
		files := index.I.List()
		log.Success("found %d files in index:", len(files))

		for _, f := range files {
			log.Info(f)
		}

	default:
		log.Warn("'%s' is not a valid command.", args[0])
		log.Info("valid commands: index, lookup <key>, delete <key>, exit")
	}

	return err
}