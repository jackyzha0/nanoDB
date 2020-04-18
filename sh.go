package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jackyzha0/nanoDB/index"
	"github.com/jackyzha0/nanoDB/log"
)

// DefaultDepth is the default depth to resolve reference to
const DefaultDepth = 0

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
	case "delete":
		return deleteWrapper(args)
	case "regenerate":
		index.I.Regenerate()
	default:
		log.Warn("'%s' is not a valid command.", args[0])
		log.Info("valid commands: index, lookup <key> <depth>, delete <key>, regenerate, exit")
	}
	return err
}

func parseDepthFromArgs(args []string) int {
	if len(args) < 3 {
		// no depth argument, use default
		return DefaultDepth
	}

	if parsedInt, err := strconv.Atoi(args[2]); err == nil {
		return parsedInt
	}

	return DefaultDepth
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

	log.Success("found key %s:", key)

	// get file bytes
	m, err := f.ToMap()
	if err != nil {
		return err
	}

	// resolve refs
	depth := parseDepthFromArgs(args)
	log.Info("resolving reference to depth %d...", depth)
	resolvedMap := index.ResolveReferences(m, depth)

	// back to bytes
	b, err := json.Marshal(resolvedMap)
	if err != nil {
		return err
	}

	// pretty format
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, b, "", "\t")
	if err != nil {
		return err
	}

	log.Info("%s", prettyJSON.String())
	return nil
}

func deleteWrapper(args []string) error {
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

	// attempt delete file
	err := index.I.Delete(f)
	if err != nil {
		return err
	}

	log.Success("deleted key %s", key)
	return nil
}
