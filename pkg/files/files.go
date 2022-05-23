package files

import (
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// PROJECT represents project in command line arguments for push, pull and yank commands.
	PROJECT = "project"
	// WORKFLOW represents workflow in command line arguments for push, pull and yank commands.
	WORKFLOW = "workflow"
	// JOB represents job in command line arguments for push, pull and yank commands.
	JOB = "job"
	// ExpirePrefix is where object expires are stored in the same bucket.
	ExpirePrefix = "var/expires-in/"
)

var (
	pluralCategory = map[string]string{
		PROJECT:  "projects",
		WORKFLOW: "workflows",
		JOB:      "jobs",
	}
	// CategoryEnv returns environment variable names for project ID, workflow ID and job ID.
	CategoryEnv = map[string]string{
		PROJECT:  "SEMAPHORE_PROJECT_ID",
		WORKFLOW: "SEMAPHORE_WORKFLOW_ID",
		JOB:      "SEMAPHORE_JOB_ID",
	}
	categoryID string
	pluralName string
)

// Returns source and destination paths to push a file to storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path storage bucket.
func PushPaths(dst, src string) (string, string) {
	newDst := ToRelative(dst)
	newDst = PrefixedPathFromSource(newDst, src)
	newSrc := path.Clean(src)

	log.Debug("Paths for pushing...\n")
	log.Debugf("> Input destination: '%s'\n", dst)
	log.Debugf("> Input source: '%s'\n", src)
	log.Debugf("> Output destination: '%s'\n", newDst)
	log.Debugf("> Output source: '%s'\n", newSrc)

	return newDst, newSrc
}

// PullPaths returns source and destination paths to pull a file from remote storage.
// Source path becomes a category prefixed path to the storage bucket,
// destination path becomes a relative path on the file system.
func PullPaths(dst, src string) (string, string) {
	newSrc := ToRelative(src)
	newDst := PathFromSource(dst, newSrc)
	newSrc = PrefixedPath(newSrc)
	newDst = path.Clean(newDst)

	log.Debug("Paths for pulling...\n")
	log.Debugf("> Input destination: %s\n", dst)
	log.Debugf("> Input source: %s\n", src)
	log.Debugf("> Output destination: %s\n", newDst)
	log.Debugf("> Output source: %s\n", newSrc)

	return newDst, newSrc
}

// Returns path to yank a file from the remote storage.
// Path becomes a category prefixed path to the storage bucket.
func YankPath(f string) string {
	newF := ToRelative(f)
	newF = PrefixedPath(newF)

	log.Debug("Paths for yanking...\n")
	log.Debugf("> Input file: %s\n", f)
	log.Debugf("> Output file: %s\n", newF)

	return newF
}

// InitPathID initiates path category ID. The default is an empty string, in that case the ID is read
// from environment variable. Otherwise it comes from command-line argument --job-id or --workflow-id.
func InitPathID(category, defVal string) error {
	if len(defVal) == 0 {
		categoryID = os.Getenv(CategoryEnv[category])
	} else {
		categoryID = defVal
	}

	if len(categoryID) == 0 {
		return fmt.Errorf("please set %sID with %s env var or related flag", category,
			CategoryEnv[category])
	}

	pluralName = pluralCategory[category]
	return nil
}

// PrefixedPath returns paths for Google Cloud Storage.
// For project files, it returns like: artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip
// For workflow files, it returns like: artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip
// For job files, it returns like: artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip
func PrefixedPath(filepath string) string {
	return path.Join("artifacts", pluralName, categoryID, filepath)
}

// PrefixedPathFromSource returns a path for Google Cloud Storage, where destination filename can be
// empty. In this case filename is gained from source filename, eg. uploading /from/this/path/x.zip
// with empty --destination to the project will return artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip,
// but with --destination=y.zip will result in artifacts/projects/<SEMAPHORE_PROJECT_ID>/y.zip .
func PrefixedPathFromSource(dstFilepath, srcFilepath string) string {
	dstFilepath = PathFromSource(dstFilepath, srcFilepath)
	return PrefixedPath(dstFilepath)
}

// PathFromSource returns a path where destination filename can be empty. If it's empty, the name is
// gained from the source filename.
func PathFromSource(dstFilepath, srcFilepath string) string {
	if len(dstFilepath) == 0 {
		dstFilepath = path.Base(srcFilepath)
	}
	return dstFilepath
}

// ToRelative removes all ./, ../ etc prefixes from the string.
func ToRelative(filepath string) string {
	cleaned := path.Clean(filepath)
	if len(cleaned) == strings.Count(cleaned, ".") {
		return ""
	}
	trimmed := strings.TrimLeft(cleaned, "./") // removed . and / chars from left
	left := cleaned[:len(cleaned)-len(trimmed)]
	// looking for . on the right side of the cut of left part
	farLeft := strings.TrimRight(left, ".")
	return cleaned[len(farLeft):]
}

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
