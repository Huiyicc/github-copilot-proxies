package static

import "embed"

var (
	//go:embed public
	Public embed.FS
)
