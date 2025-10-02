package utils

import (
	"path/filepath"
	"strings"
)

func TruncatePath(fullPath string, levels int) string {
	if levels <= 0 {
		return ""
	}

	parts := strings.Split(fullPath, string(filepath.Separator))

	if len(parts) <= levels {
		return fullPath
	}

	start := len(parts) - levels
	return filepath.Join(parts[start:]...)
}
