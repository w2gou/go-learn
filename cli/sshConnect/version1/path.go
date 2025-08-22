package version1

import (
	"os"
	"strings"
)

func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = home + path[1:]
	}
	return path
}
