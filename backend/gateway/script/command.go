package script

import (
	"log"
	"os"
	"strings"
)

func Commands() bool {
	var scriptName string

	run := false
	scriptFlag := false

	for _, arg := range os.Args[1:] {
		if arg == "--run" {
			run = true
		}
		if strings.HasPrefix(arg, "--script:") {
			scriptFlag = true
			scriptName = strings.TrimPrefix(arg, "--script:")
		}
	}

	if scriptFlag {
		if err := Script(scriptName); err != nil {
			log.Fatalf("error script: %v", err)
		}
		log.Println("script run successfully")
	}

	if run {
		return true
	}

	return false
}
