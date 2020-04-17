package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackyzha0/nanoDB/index"
	"github.com/jackyzha0/nanoDB/log"
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
		if err = execInput(input, dir); err != nil {
			log.Warn("err executing input: %s", err.Error())
		}
	}
}

func execInput(input string, dir string) (err error) {
	input = strings.TrimSuffix(input, "\n")
	args := strings.Split(input, " ")

	switch args[0] {
	case "index":
		indexWrapper()
	case "exit":
		cleanup(dir)
		os.Exit(0)
	case "lookup":
		return lookupWrapper(args)
	default:
		log.Warn("'%s' is not a valid command.", args[0])
		log.Info("valid commands: index, lookup <key>, delete <key>, exit")
	}
	return err
}

func indexWrapper() {
	files := index.I.List()
	log.Success("found %d files in index:", len(files))

	for _, f := range files {
		log.Info(f)
	}
}

func lookupWrapper(args []string) error {
	// assert theres a key
	if len(args) < 2 {
		err := fmt.Errorf("no key provided")
		return err
	}

	key := args[1]

	// lookup key, return err if not found
	f, ok := index.I.Lookup(key)
	if !ok {
		err := fmt.Errorf("key doesn't exist")
		return err
	}

	// get file bytes
	b, err := f.GetByteArray()
	if err != nil {
		return err
	}

	// pretty format
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, b, "", "\t")
	if err != nil {
		return err
	}

	log.Success("found key %s:", key)
	log.Info("%s", prettyJSON.String())
	return nil
}