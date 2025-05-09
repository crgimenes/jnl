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

func configHome() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get home directory:", err)
		}
		configHome = filepath.Join(home, ".config")
	}
	return configHome
}

func getInitLuaPath() string {
	configHome := configHome()
	return filepath.Join(configHome, "jnl", "init.lua")
}

// createConfigDir creates the config directory if it does not exist.
func createConfigDir() {
	configHome := configHome()
	configDir := filepath.Join(configHome, "jnl")
	_, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(configDir, 0700)
			if err != nil {
				log.Fatal("Failed to create config directory:", err)
			}
			return
		}
		log.Fatal("Failed to check config directory:", err)
	}
}

func runLuaFile(name string) {
	if !fileExists(name) {
		// TODO: load defaults
		return
	}

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

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	createConfigDir()
	initFile := getInitLuaPath()

	cmd := "add"
	// check first parameter
	if len(os.Args) > 2 {
		cmd = os.Args[1]
		return
	}

	if fileExists("./jnl_init.lua") {
		initFile = "./jnl_init.lua"
	}

	runLuaFile(initFile)

	switch cmd {
	case "add":
		if isTTY {
			// open $EDITOR with a temporary file and use the file as the content
			return
		}

		// load the content from stdin using readAll and it as the content
		return
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
