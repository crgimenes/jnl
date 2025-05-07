package main

import (
	"fmt"
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
		return
	case "rm":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a note to remove.")
			return
		}
		// remove the note
		return
	case "edit":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a note to edit.")
			return
		}
		// open the note in $EDITOR
		return
	case "less":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a note to view.")
			return
		}
		// view the note in less
		return
	}

}
