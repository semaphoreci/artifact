package files

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// isFile returns if the given path points to a file in the local file system.
func isFile(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		return !fi.IsDir(), nil
	}

	return false, fmt.Errorf("error finding file '%s': %v", filename, err)
}

// isDir returns if the given path points to a directory in the local file system.
func isDir(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		return fi.IsDir(), nil
	}

	return false, fmt.Errorf("error finding directory '%s': %v", filename, err)
}

// Checks if the given source exists and is a file.
func IsFileSrc(src string) (bool, error) {
	isFile, err := isFile(src)
	if err != nil {
		return false, err
	}

	if isFile {
		return true, nil
	}

	isDir, err := isDir(src)
	if err != nil {
		return false, err
	}

	if isDir {
		return false, nil
	}

	return false, fmt.Errorf("path '%s' doesn't exist", src)
}

// Removes all ./, ../ etc prefixes from the string.
func ToRelative(filepath string) string {
	cleaned := path.Clean(filepath)
	if len(cleaned) == strings.Count(cleaned, ".") {
		return ""
	}

	// removed . and / chars from left
	trimmed := strings.TrimLeft(cleaned, "./")
	left := cleaned[:len(cleaned)-len(trimmed)]

	// looking for . on the right side of the cut of left part
	farLeft := strings.TrimRight(left, ".")
	return cleaned[len(farLeft):]
}
