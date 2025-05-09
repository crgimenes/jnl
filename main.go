package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/crgimenes/jnl/lua"
	"golang.org/x/term"
)

var (
	isTTY      = term.IsTerminal(int(os.Stdout.Fd()))
	jornalPath string
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

func simpleSlugify(s string) string {
	var b strings.Builder
	b.Grow(len(s)) // preallocate for performance
	prevDash := false

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}

	slug := b.String()
	slug = strings.Trim(slug, "-")
	return slug
}

func journalFilename(content []byte) string {
	// extract first line
	idx := bytes.IndexByte(content, '\n')
	firstLine := string(content)
	if idx >= 0 {
		firstLine = string(content[:idx])
	}
	firstLine = strings.TrimSpace(firstLine)

	// slugify
	slug := simpleSlugify(firstLine)

	// current timestamp
	ts := time.Now().Format("20060102T150405")

	if slug == "" {
		return ts + ".md"
	}
	return fmt.Sprintf("%s-%s.md", ts, slug)
}

func runLuaFile(name string) {
	jornalPath = "./" // TODO: get better default path

	if !fileExists(name) {
		return
	}

	// Create a new Lua state.
	L := lua.New()
	defer L.Close()

	L.SetGlobal("jornal_path", jornalPath)

	// Read the Lua file.
	b, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		log.Fatal(err)
	}

	err = L.DoString(string(b))
	if err != nil {
		log.Fatal(err)
	}

	jornalPath = L.MustGetString("jornal_path")
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
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			tmpFile, err := os.CreateTemp("", "jnl-*.md")
			if err != nil {
				log.Fatal("Failed to create temporary file:", err)
			}
			defer tmpFile.Close()
			defer os.Remove(tmpFile.Name())

			// open the file with the editor using exec.Command
			cmd := exec.Command(editor, tmpFile.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				log.Fatal("Failed to run editor:", err)
			}

			// read the content of the file
			content, err := os.ReadFile(tmpFile.Name())
			if err != nil {
				log.Fatal("Failed to read temporary file:", err)
			}

			if len(content) == 0 {
				// enpty file, do not save
				return
			}

			// save the content to the journal path
			journalFile := filepath.Join(jornalPath, journalFilename(content))
			log.Println("Saving journal entry to:", journalFile)

			err = os.WriteFile(journalFile, content, 0600)
			if err != nil {
				log.Fatal("Failed to write journal file:", err)
			}

			fmt.Println("Journal entry saved to:", journalFile)
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
