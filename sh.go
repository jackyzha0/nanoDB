package main

import (
	"bufio"
	"os"
	"strings"

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
		if err = execInput(input); err != nil {
			log.Warn("err executing input: %s", err.Error())
		}
	}
}

func execInput(input string) error {
	args := strings.Split(input, " ")
	log.Info("%+v", args)
	return nil
}