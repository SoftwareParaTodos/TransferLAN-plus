package transfer

import (
	"path/filepath"
	"strings"
)

func SafeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "transferlan_file"
	}
	name = strings.ReplaceAll(name, "\x00", "")
	return name
}
