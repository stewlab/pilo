package main

import (
	"embed"
	"pilo/cmd/pilo"
)

//go:embed all:flake
var flakeFS embed.FS

func main() {
	pilo.Execute(flakeFS)
}
