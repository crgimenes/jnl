package main

import (
	"fmt"
	"strings"
	"time"
)

func Example_parseHeader() {
	info := map[string]string{
		"date":   time.Now().Format(time.RFC3339),
		"dir":    "~/Projects/foo",
		"branch": "feature-x",
		"tags":   "@work, @diary",
	}

	// Build the header and create a Markdown entry.
	entry := buildHeader(info) + "## Today's notes\n\nHello world!\n"

	// Parse the header back.
	hdr, body := parseHeader([]byte(entry))

	fmt.Println(hdr["dir"])
	fmt.Println(strings.TrimSpace(string(body)))
	// Output:
	// ~/Projects/foo
	// ## Today's notes
	//
	// Hello world!
}
