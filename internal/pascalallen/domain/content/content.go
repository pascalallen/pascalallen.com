package content

import (
	"embed"
	"io/fs"
)

//go:embed home
var home embed.FS

func FS() fs.FS {
	sub, err := fs.Sub(home, "home")
	if err != nil {
		panic(err)
	}
	return sub
}
