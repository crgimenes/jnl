package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/crgimenes/jnl/lua"
	"golang.org/x/term"
)

var (
	isTTY = term.IsTerminal(int(os.Stdout.Fd()))
)

func getInitLuaPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get home directory:", err)
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "jnl", "init.lua")
}

func runLuaFile(name string) {
	// Create a new Lua state.
	L := lua.New()
	defer L.Close()

	// Read the Lua file.
	b, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		log.Fatal(err)
	}

	err = L.DoString(string(b))
	if err != nil {
		log.Fatal(err)
	}
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// check first parameter
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command to run.")
		return
	}

	cmd := os.Args[1]

	initFile := getInitLuaPath()
	if fileExists(initFile) {
		runLuaFile(initFile)
	} else {
		// load local config
		initFile = "./jnl_init.lua"
		runLuaFile(initFile)
	}

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
		log.Println("not implemented")
		return
	case "review":
		log.Println("not implemented")
		return
	case "help":
		log.Println("not implemented")
		return
	case "version":
		log.Println("not implemented")
		return
	case "config":
		log.Println("not implemented")
		return
	case "init":
		log.Println("not implemented")
		return
	case "serve": // start a web server
		log.Println("not implemented")
		return
	case "sync": // sync with remote server
		log.Println("not implemented")
		return
	default:
		fmt.Println("Unknown command:", cmd)
	}
}
