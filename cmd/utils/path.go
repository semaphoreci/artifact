package utils

import (
	"os"
	"path"
	"strings"
)

// PROJECT represents project in command line arguments for push, pull and yank commands.
const PROJECT = "project"

// WORKFLOW represents workflow in command line arguments for push, pull and yank commands.
const WORKFLOW = "workflow"

// JOB represents job in command line arguments for push, pull and yank commands.
const JOB = "job"

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
)

// PrefixedPath returns paths for Google Cloud Storage.
// For project files, it returns like: /artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip
// For workflow files, it returns like: /artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip
// For job files, it returns like: /artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip
func PrefixedPath(category, filepath string) string {
	pluralName := pluralCategory[category]
	categoryID := os.Getenv(CategoryEnv[category])
	return path.Join("/artifacts", pluralName, categoryID, filepath)
}

// PrefixedPathFromSource returns a path for Google Cloud Storage, where destination filename can be
// empty. In this case filename is gained from source filename, eg. uploading /from/this/path/x.zip
// with empty --destination to the project will return /artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip,
// but with --destination=y.zip will result in /artifacts/projects/<SEMAPHORE_PROJECT_ID>/y.zip .
func PrefixedPathFromSource(category, dstFilepath, srcFilepath string) string {
	dstFilepath = PathFromSource(dstFilepath, srcFilepath)
	return PrefixedPath(category, dstFilepath)
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
