package main

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

// buildHeader creates the YAML frontmatter header block.
func buildHeader(info map[string]string) string {
	var b strings.Builder

	v, ok := info["tag"]
	if !ok || v == "" || v == tagPrefix {
		// remove tag if it is empty or equals to the default tagPrefix
		delete(info, "tag")
	}

	sortedKeys := make([]string, 0, len(info))
	for k := range info {
		sortedKeys = append(sortedKeys, k)
	}
	// Sort keys to ensure consistent order in the Header
	sort.Strings(sortedKeys)

	b.WriteString("---\n") // opening delimiter
	for _, k := range sortedKeys {
		v := info[k]
		if k == "tags" {
			writeYAMLTagList(&b, v)
			continue
		}
		if yamlNeedsQuoting(v) {
			fmt.Fprintf(&b, "%s: \"%s\"\n", k, strings.ReplaceAll(v, "\"", "\\\""))
		} else {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	b.WriteString("---\n\n") // closing delimiter + blank line
	return b.String()
}

// writeYAMLTagList writes the tags key as a YAML block sequence.
// Input is a comma-separated string of tags (with or without prefix).
// Output is Obsidian-compatible:
//
//	tags:
//	  - tag1
//	  - tag2
func writeYAMLTagList(b *strings.Builder, tags string) {
	parts := strings.Split(tags, ",")
	var clean []string
	for _, t := range parts {
		t = strings.TrimSpace(t)
		t = strings.TrimPrefix(t, tagPrefix) // strip legacy @ if present
		if t != "" {
			clean = append(clean, t)
		}
	}
	if len(clean) == 0 {
		return
	}
	b.WriteString("tags:\n")
	for _, t := range clean {
		fmt.Fprintf(b, "  - %s\n", t)
	}
}

// yamlNeedsQuoting reports whether v must be double-quoted in YAML output.
func yamlNeedsQuoting(v string) bool {
	if v == "" {
		return true
	}
	// Characters that are reserved or problematic at the start of a YAML
	// plain scalar (see YAML 1.2 spec: c-indicator).
	switch v[0] {
	case '@', '`', '&', '*', '!', '{', '[', '>', '|', '\'', '"', '%', '#', '?', ',':
		return true
	}
	// Values containing : followed by space, or trailing/leading spaces.
	if strings.Contains(v, ": ") || strings.Contains(v, " #") {
		return true
	}
	if v[0] == ' ' || v[len(v)-1] == ' ' {
		return true
	}
	return false
}

// parseHeader extracts the header and returns (header, body).
// It supports both YAML frontmatter (---) and legacy Filo-style (;;; jnl)
// delimiters for backward compatibility.
func parseHeader(md []byte) (map[string]string, []byte) {
	sc := bufio.NewScanner(bytes.NewReader(md))

	header := make(map[string]string)
	var body bytes.Buffer

	// Peek at the first non-empty line to determine format.
	var firstLine string
	for sc.Scan() {
		firstLine = strings.TrimSpace(sc.Text())
		if firstLine != "" {
			break
		}
		body.WriteString(sc.Text() + "\n")
	}

	switch {
	case firstLine == "---":
		return parseYAMLHeader(sc, &body)
	case strings.HasPrefix(firstLine, ";;;"):
		return parseLegacyHeader(sc, &body)
	default:
		// No header found; the first line is part of the body.
		body.WriteString(firstLine + "\n")
		for sc.Scan() {
			body.WriteString(sc.Text() + "\n")
		}
		return header, body.Bytes()
	}
}

// parseYAMLHeader parses a YAML frontmatter block delimited by ---.
// The opening --- has already been consumed by the caller.
func parseYAMLHeader(sc *bufio.Scanner, body *bytes.Buffer) (map[string]string, []byte) {
	var raw bytes.Buffer
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		raw.WriteString(line + "\n")
	}

	// Collect remaining body.
	for sc.Scan() {
		body.WriteString(sc.Text() + "\n")
	}

	header := make(map[string]string)
	var parsed map[string]any
	if err := yaml.Unmarshal(raw.Bytes(), &parsed); err != nil {
		return header, body.Bytes()
	}

	for k, v := range parsed {
		header[k] = yamlValueToString(v)
	}
	return header, body.Bytes()
}

// yamlValueToString converts a YAML value to a string representation
// compatible with the existing key:value header format.
// Arrays (e.g., tags: [a, b]) are joined with ", ".
func yamlValueToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []any:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(v)
	}
}

// parseLegacyHeader parses the legacy ;;; jnl / ;;; header format.
// The opening ;;; jnl line has already been consumed by the caller.
func parseLegacyHeader(sc *bufio.Scanner, body *bytes.Buffer) (map[string]string, []byte) {
	header := make(map[string]string)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, ";;;") {
			break
		}
		kv := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(kv) == 2 {
			header[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	// Collect remaining body.
	for sc.Scan() {
		body.WriteString(sc.Text() + "\n")
	}
	return header, body.Bytes()
}
