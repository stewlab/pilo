package pilo

import (
	"embed"
	"pilo/internal/cli"
)

func Execute(flakeFS embed.FS) {
	cli.Execute(flakeFS)
}
