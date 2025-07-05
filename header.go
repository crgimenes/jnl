package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// buildHeader creates the `;;;` header block.
func buildHeader(info map[string]string) string {
	var b strings.Builder

	b.WriteString(";;; jnl\n") // opening delimiter
	for k, v := range info {
		fmt.Fprintf(&b, "%s: %s\n", k, v)
	}
	b.WriteString(";;;\n\n") // closing delimiter + blank line
	return b.String()
}

// parseHeader extracts the header and returns (header, body).
func parseHeader(md []byte) (map[string]string, []byte) {
	sc := bufio.NewScanner(bytes.NewReader(md))

	header := make(map[string]string)
	var body bytes.Buffer
	inHdr := false

	for sc.Scan() {
		line := sc.Text()

		switch {
		case !inHdr && strings.HasPrefix(line, ";;; jnl"):
			inHdr = true // start header
			continue
		case inHdr && strings.HasPrefix(line, ";;;"):
			inHdr = false // end header
			continue
		}

		if inHdr {
			kv := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(kv) == 2 {
				header[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
			continue
		}
		body.WriteString(line + "\n")
	}
	return header, body.Bytes()
}
