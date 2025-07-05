package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/crgimenes/jnl/lua"
	glua "github.com/yuin/gopher-lua"
	"golang.org/x/term"
)

var (
	GitTag             = "v0.0.0"
	force              = false // depends on the command
	isTTY              = term.IsTerminal(int(os.Stdout.Fd()))
	journalPath        = "./"
	journalTitle       = "" // optional title for the journal entry (first line if set)
	journalTitlePrefix = ""
	tagPrefix          = "@"
	listenAddr         = ":8080" // default address for the web server

	// Create a new Lua state.
	L = lua.New()
)

func preProc(text string) string {
	ls := L.GetState()
	fn := ls.GetGlobal("PreProc")
	if _, ok := fn.(*glua.LFunction); !ok {
		return text
	}
	err := ls.CallByParam(glua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, glua.LString(text))
	if err != nil {
		return text
	}

	ret := ls.Get(-1)
	ls.Pop(1)
	s, ok := ret.(glua.LString)
	if ok {
		return string(s)
	}
	return text
}

func postProc(text string) string {
	ls := L.GetState()
	fn := ls.GetGlobal("PostProc")
	if _, ok := fn.(*glua.LFunction); !ok {
		return text
	}
	err := ls.CallByParam(glua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, glua.LString(text))
	if err != nil {
		return text
	}

	ret := ls.Get(-1)
	ls.Pop(1)
	s, ok := ret.(glua.LString)
	if ok {
		return string(s)
	}
	return text
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
	// Parse header to extract title or first line of body
	header, body := parseHeader(content)

	var firstLine string

	// Check if there's a title in the header
	if title, hasTitle := header["title"]; hasTitle && title != "" {
		firstLine = title
	} else {
		// Extract first meaningful line from body (skip empty lines)
		lines := bytes.SplitSeq(body, []byte("\n"))
		for line := range lines {
			lineStr := strings.TrimSpace(string(line))
			if lineStr != "" {
				firstLine = lineStr
				break
			}
		}

		// Remove markdown headers (# ## ###, etc.)
		firstLine = strings.TrimSpace(strings.TrimLeft(firstLine, "#"))
		firstLine = strings.TrimSpace(firstLine)
	}

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
	if !fileExists(name) {
		return
	}

	L.SetGlobal("JournalPath", journalPath)
	L.SetGlobal("JournalTitlePrefix", journalTitlePrefix)
	L.SetGlobal("TagPrefix", tagPrefix)
	L.SetGlobal("ListenAddr", listenAddr)

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

	journalTitlePrefix = L.MustGetString("JournalTitlePrefix")
	tagPrefix = L.MustGetString("TagPrefix") // default to @
	listenAddr = L.MustGetString("ListenAddr")

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
		fmt.Println("No journal entries found.")
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

// isInGitRepository checks if the current directory or any parent directory contains a .git folder
func isInGitRepository() bool {
	currentDir, err := os.Getwd()
	if err != nil {
		return false
	}

	// Walk up the directory tree looking for .git
	for {
		gitDir := filepath.Join(currentDir, ".git")
		if info, err := os.Stat(gitDir); err == nil {
			// Check if it's a directory or a file (for git worktrees)
			if info.IsDir() {
				return true
			}
			// For git worktrees, .git is a file containing the path to the real .git directory
			if info.Mode().IsRegular() {
				return true
			}
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We've reached the root directory
			break
		}
		currentDir = parentDir
	}

	return false
}

func getGitBranch() (string, bool) {
	// First check if we're in a git repository
	if !isInGitRepository() {
		return "", false
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		// Even if we found .git, the command might fail (e.g., no commits yet)
		return "", false
	}

	branch := string(bytes.TrimSpace(out))
	// Handle detached HEAD state
	if branch == "HEAD" {
		// Try to get the commit hash instead
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
		cmd.Env = os.Environ()
		out, err = cmd.Output()
		if err != nil {
			return "", false
		}
		return "detached-" + string(bytes.TrimSpace(out)), true
	}

	return branch, true
}

func webServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "drafts")
	})

	s := &http.Server{
		Handler:        mux,
		Addr:           listenAddr,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server on %s\n", listenAddr)
	log.Fatal(s.ListenAndServe())
}

// editJournalEntry edits an existing journal entry
func editJournalEntry(filename string) error {
	// Construct full path
	journalFile := filepath.Join(journalPath, filename)
	if !strings.HasSuffix(journalFile, ".md") {
		journalFile += ".md"
	}

	// Check if file exists
	if !fileExists(journalFile) {
		return fmt.Errorf("journal entry not found: %s", filename)
	}

	// Read existing content
	content, err := os.ReadFile(journalFile)
	if err != nil {
		return fmt.Errorf("failed to read journal entry: %v", err)
	}

	// Parse existing header and body
	header, body := parseHeader(content)

	// Open editor with temporary file
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpFile, err := os.CreateTemp("", "jnl-edit-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	// Write current content to temp file
	reconstructed := buildHeader(header) + string(body)
	_, err = tmpFile.WriteString(reconstructed)
	if err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}

	// Open editor
	ls := L.GetState()
	raw := ls.GetGlobal("Exec")
	fn, ok := raw.(*glua.LFunction)
	if ok {
		err := ls.CallByParam(glua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}, glua.LString(editor), glua.LString(tmpFile.Name()))
		if err != nil {
			return fmt.Errorf("lua Exec error: %v", err)
		}
	} else {
		cmd := exec.Command(editor, tmpFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run editor: %v", err)
		}
	}

	// Read edited content
	editedContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited file: %v", err)
	}

	editedContent = []byte(postProc(string(editedContent)))

	// Check if content changed
	if bytes.Equal(content, editedContent) && !force {
		fmt.Println("No changes detected, not saving. Use --force to save anyway.")
		return nil
	}

	// Save the updated content
	err = os.WriteFile(journalFile, editedContent, 0600)
	if err != nil {
		return fmt.Errorf("failed to save journal entry: %v", err)
	}

	fmt.Println("Journal entry updated:", journalFile)
	return nil
}

// catJournalEntry displays the content of a journal entry
func catJournalEntry(filename string) error {
	// Construct full path
	journalFile := filepath.Join(journalPath, filename)
	if !strings.HasSuffix(journalFile, ".md") {
		journalFile += ".md"
	}

	// Check if file exists
	if !fileExists(journalFile) {
		return fmt.Errorf("journal entry not found: %s", filename)
	}

	// Read and display content
	content, err := os.ReadFile(journalFile)
	if err != nil {
		return fmt.Errorf("failed to read journal entry: %v", err)
	}

	fmt.Print(string(content))
	return nil
}

// showJournalEntry displays a journal entry with header information separated
func showJournalEntry(filename string) error {
	// Construct full path
	journalFile := filepath.Join(journalPath, filename)
	if !strings.HasSuffix(journalFile, ".md") {
		journalFile += ".md"
	}

	// Check if file exists
	if !fileExists(journalFile) {
		return fmt.Errorf("journal entry not found: %s", filename)
	}

	// Read content
	content, err := os.ReadFile(journalFile)
	if err != nil {
		return fmt.Errorf("failed to read journal entry: %v", err)
	}

	// Parse header and body
	header, body := parseHeader(content)

	// Display header information
	fmt.Printf("=== %s ===\n", filename)
	if len(header) > 0 {
		fmt.Println("Header:")
		for k, v := range header {
			fmt.Printf("  %s: %s\n", k, v)
		}
		fmt.Println()
	}

	// Display body
	fmt.Print(string(body))
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	createConfigDir()
	initFile := getInitLuaPath()

	cmd := "add"
	// parse global options
	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			if strings.HasPrefix(arg, "-") {
				if arg == "--title" || arg == "-t" {
					i++
					if i >= len(os.Args) {
						log.Fatal("Missing title argument")
					}
					journalTitle = os.Args[i]
					continue
				}

				if arg == "--force" || arg == "-f" {
					force = true
					continue
				}

				continue
			}
			cmd = arg
			break
		}
	}

	if fileExists("./jnl_init.lua") {
		initFile = "./jnl_init.lua"
	}

	runLuaFile(initFile)

	defer func() {
		defer L.Close()
	}()

	switch cmd {
	case "add":
		// parse arguments
		// --commit (run `exec.Command("git", "commit", "-F", journalFile).Run()` after saving)
		runCommitAfterSave := false
		if len(os.Args) > 1 {
			for i := 1; i < len(os.Args); i++ {
				arg := os.Args[i]
				if arg == "--commit" || arg == "-c" {
					runCommitAfterSave = true
					continue
				}
			}
		}

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

		// Build header info
		headerInfo := make(map[string]string)

		// Add date
		headerInfo["date"] = time.Now().Format(time.RFC3339)

		// get current directory info
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal("Failed to get current directory:", err)
		}
		wd, err = filepath.Abs(wd)
		if err != nil {
			log.Fatal("Failed to get absolute path:", err)
		}
		// remove home directory from path
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get home directory:", err)
		}
		wd = strings.Replace(wd, home, "", 1)
		wd = strings.TrimPrefix(wd, "/")
		tagArray := strings.Split(wd, "/")

		beautifiedPath := "~/" + strings.TrimPrefix(wd, "/")
		headerInfo["dir"] = beautifiedPath

		// Add user
		userName := os.Getenv("USER")
		if userName == "" {
			userName = os.Getenv("USERNAME")
		}
		if userName != "" {
			headerInfo["user"] = userName
		}

		// get git branch name
		gitBranch, ok := getGitBranch()
		if ok {
			headerInfo["branch"] = gitBranch
			tagArray = append(tagArray, gitBranch)
		}

		// add @ in front of each tag
		for i := range tagArray {
			tagArray[i] = "@" + tagArray[i]
		}

		tags := strings.Join(tagArray, ", ")
		headerInfo["tags"] = tags

		// Add title if provided
		if journalTitle != "" {
			headerInfo["title"] = journalTitle
		}

		// Build the header using the new function
		s := buildHeader(headerInfo)
		s = preProc(s)

		prevContent := s

		// write the content to the temporary file
		_, err = tmpFile.WriteString(s)
		if err != nil {
			log.Fatal("Failed to write to temporary file:", err)
		}

		/*
			// open the file with the editor using exec.Command
			cmd := exec.Command(editor, tmpFile.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = os.Environ()
			err = cmd.Run()
			if err != nil {
				log.Fatal("Failed to run editor:", err)
			}
		*/

		// obtém a LState do gopher-lua
		ls := L.GetState()
		// tenta recuperar a função Exec do Lua
		raw := ls.GetGlobal("Exec")
		fn, ok := raw.(*glua.LFunction)
		if ok {
			// chama Exec(editor, tmpFile)
			err := ls.CallByParam(glua.P{
				Fn:      fn,
				NRet:    0,
				Protect: true,
			}, glua.LString(editor), glua.LString(tmpFile.Name()))
			if err != nil {
				log.Fatalf("Lua Exec error: %v", err)
			}
		} else {
			// fallback padrão: chamar o binário diretamente
			cmd := exec.Command(editor, tmpFile.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = os.Environ()
			err := cmd.Run()
			if err != nil {
				log.Fatal("Failed to run editor:", err)
			}
		}

		// read the content of the file
		content, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			log.Fatal("Failed to read temporary file:", err)
		}

		content = []byte(postProc(string(content)))

		// check if the content is empty
		if len(content) == 0 {
			fmt.Println("Empty file, not saving.")
			return
		}

		if bytes.Equal(content, []byte(prevContent)) &&
			journalTitle == "" && !force {
			fmt.Println("Aparently no changes, not saving, use --force to save.")
			return
		}

		// save the content to the journal path
		journalFile := filepath.Join(journalPath, journalFilename(content))

		err = os.WriteFile(journalFile, content, 0600)
		if err != nil {
			log.Fatal("Failed to write journal file:", err)
		}

		fmt.Println("Journal entry saved to:", journalFile)
		if runCommitAfterSave {
			fmt.Println("git commit...")
			cmd := exec.Command("git", "commit", "-F", journalFile)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = os.Environ()
			err = cmd.Run()
			if err != nil {
				log.Fatalf("Failed to run `git commit -F %s`: %v\n", journalFile, err)
			}
		}

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
	case "path":
		fmt.Println(journalPath)
		return
	case "rm":
		log.Println("not implemented")
		return
	case "edit":
		// Edit an existing journal entry
		if len(os.Args) < 3 {
			log.Fatal("Usage: jnl edit <filename>")
		}
		filename := os.Args[2]
		err := editJournalEntry(filename)
		if err != nil {
			log.Fatal(err)
		}
		return
	case "less":
		// Show a journal entry with header information
		if len(os.Args) < 3 {
			log.Fatal("Usage: jnl less <filename>")
		}
		filename := os.Args[2]
		err := showJournalEntry(filename)
		if err != nil {
			log.Fatal(err)
		}
		return
	case "cat":
		// Display raw content of a journal entry
		if len(os.Args) < 3 {
			log.Fatal("Usage: jnl cat <filename>")
		}
		filename := os.Args[2]
		err := catJournalEntry(filename)
		if err != nil {
			log.Fatal(err)
		}
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
		webServer()
		return
	case "sync": // sync with remote server
		log.Println("not implemented")
		return
	default:
		fmt.Println("Unknown command:", cmd)
	}
}
