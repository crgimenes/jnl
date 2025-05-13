package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/crgimenes/jnl/lua"
	"golang.org/x/term"
)

var (
	GitTag      = "v0.0.0"
	isTTY       = term.IsTerminal(int(os.Stdout.Fd()))
	journalPath string
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
	ts := time.Now().Format("2006-01-02T15-04-05")

	if slug == "" {
		return ts + ".md"
	}
	return fmt.Sprintf("%s-%s.md", ts, slug)
}

func runLuaFile(name string) {
	journalPath = "./"

	if !fileExists(name) {
		return
	}

	// Create a new Lua state.
	L := lua.New()
	defer L.Close()

	L.SetGlobal("JournalPath", journalPath)

	// Read the Lua file.
	b, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		log.Fatal(err)
	}

	err = L.DoString(string(b))
	if err != nil {
		log.Fatal(err)
	}

	journalPath = L.MustGetString("JournalPath")

	// resolve ~/ to full path
	if strings.HasPrefix(journalPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get home directory:", err)
		}
		journalPath = strings.Replace(journalPath, "~", home, 1)
	}

	journalPath, err = filepath.Abs(journalPath)
	if err != nil {
		log.Fatal("Failed to get absolute path:", err)
	}
}

func listJournalEntries(pattern string, showFullPath bool) error {
	// Ensure journal directory exists
	if _, err := os.Stat(journalPath); os.IsNotExist(err) {
		return fmt.Errorf("journal directory does not exist: %s", journalPath)
	}

	// Read all files from the journal directory
	files, err := os.ReadDir(journalPath)
	if err != nil {
		return fmt.Errorf("failed to read journal directory: %v", err)
	}

	// Filter and sort the files
	var matchedFiles []string
	for _, file := range files {
		// Skip directories and non-markdown files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Apply pattern filter if provided
		if pattern != "" {
			matched, err := filepath.Match(pattern, file.Name())
			if err != nil {
				// If the pattern is invalid, try as substring
				if strings.Contains(file.Name(), pattern) {
					matchedFiles = append(matchedFiles, file.Name())
				}
				continue
			}
			if matched {
				matchedFiles = append(matchedFiles, file.Name())
			}
			continue
		}
		matchedFiles = append(matchedFiles, file.Name())
	}

	// Sort files alphabetically
	sort.Strings(matchedFiles)

	if len(matchedFiles) == 0 {
		fmt.Println("Nenhuma entrada no diário encontrada.")
		return nil
	}

	for _, fileName := range matchedFiles {
		if showFullPath {
			fmt.Println(filepath.Join(journalPath, fileName))
			continue
		}
		fmt.Println(fileName)
	}

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	createConfigDir()
	initFile := getInitLuaPath()

	cmd := "add"
	// check first parameter
	if len(os.Args) > 1 {
		cmd = os.Args[1]
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
			journalFile := filepath.Join(journalPath, journalFilename(content))

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
		// List journal entries, optionally filtered by pattern
		var pattern string
		var showFullPath bool

		// Parse arguments
		for i := 2; i < len(os.Args); i++ {
			arg := os.Args[i]
			if arg == "--full-path" || arg == "-f" {
				showFullPath = true
				continue
			}
			// First non-flag argument is the pattern
			if pattern == "" && !strings.HasPrefix(arg, "-") {
				pattern = arg
			}
		}

		err := listJournalEntries(pattern, showFullPath)
		if err != nil {
			log.Fatal(err)
		}
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
		fmt.Printf("journal %s\n", GitTag)
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
