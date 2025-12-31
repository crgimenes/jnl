module github.com/crgimenes/jnl

go 1.25

require (
	github.com/crgimenes/devengine v0.0.0
	golang.org/x/term v0.38.0
)

require golang.org/x/sys v0.39.0 // indirect

replace github.com/crgimenes/devengine => ../devengine
