package utils

import (
	"os"
	"path"
	"strings"
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

// InitPathID initiates path category ID. The default is an empty string, in that case the ID is read
// from environment variable. Otherwise it comes from command-line argument --job-id or --workflow-id.
func InitPathID(category, defVal string) {
	if len(defVal) == 0 {
		categoryID = os.Getenv(CategoryEnv[category])
	} else {
		categoryID = defVal
	}
	pluralName = pluralCategory[category]
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
	return strings.TrimLeft(path.Clean(filepath), "./")
}
