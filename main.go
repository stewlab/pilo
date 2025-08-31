package main

import (
	"embed"
	"pilo/cmd/pilo"
)

//go:embed all:flake
var flakeFS embed.FS

var version = "0.0.1"

func main() {
	pilo.Execute(flakeFS, version)
}
