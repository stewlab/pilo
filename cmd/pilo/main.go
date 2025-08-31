package pilo

import (
	"embed"
	"pilo/internal/cli"
)

func Execute(flakeFS embed.FS, version string) {
	cli.Execute(flakeFS, version)
}
