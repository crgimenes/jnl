package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/term"
)

var (
	isTTY = term.IsTerminal(int(os.Stdout.Fd()))
)

func main() {
	// check first parameter
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command to run.")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "add":
		if len(os.Args) < 3 {
			if isTTY {
				// open $EDITOR with a temporary file and use the file as the content
				return
			}

			// load the content from stdin using readAll and it as the content
			return
		}
	case "ls":
		// list all the notes
		log.Println("not implemented")
		return
	case "rm":
		log.Println("not implemented")
		return
	case "edit":
		log.Println("not implemented")
		return
	case "less":
		log.Println("not implemented")
		return
	case "cat":
		log.Println("not implemented")
		return
	case "publish":
		// view the note in cat
		return
	case "review":
		// view the note in cat
		return
	default:
		fmt.Println("Unknown command:", cmd)
	}
}
