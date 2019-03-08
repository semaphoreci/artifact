package utils

import (
	"os"
	"path"
	"path/filepath"
)

const PROJECT = "project"
const WORKFLOW = "workflow"
const JOB = "job"

var (
	pluralCategory = map[string]string{
		PROJECT:  "projects",
		WORKFLOW: "workflows",
		JOB:      "jobs",
	}
	categoryEnv = map[string]string{
		PROJECT:  "SEMAPHORE_PROJECT_ID",
		WORKFLOW: "SEMAPHORE_WORKFLOW_ID",
		JOB:      "SEMAPHORE_JOB_ID",
	}
)

// prefixedPath returns paths for Google Cloud Storage.
// For project files, it returns like: /artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip
// For workflow files, it returns like: /artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip
// For job files, it returns like: /artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip
func prefixedPath(category, filepath string) string {
	pluralName := pluralCategory[category]
	categoryId := os.Getenv(categoryEnv[category])
	return path.Join("/artifacts", pluralName, categoryId, filepath)
}

// prefixedPathFromSource returns a path for Google Cloud Storage, where destination filename can be
// empty. In this case filename is gained from source filename, eg. uploading /from/this/path/x.zip
// with empty --destination to the project will return /artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip,
// but with --destination=y.zip will result in /artifacts/projects/<SEMAPHORE_PROJECT_ID>/y.zip .
func prefixedPathFromSource(category, dstFilepath, srcFilepath string) string {
	dstFilepath = pathFromSource(dstFilepath, srcFilepath)
	return prefixedPath(category, dstFilepath)
}

// pathFromSource returns a path where destination filename can be empty. If it's empty, the name is
// gained from the source filename.
func pathFromSource(dstFilepath, srcFilepath string) string {
	if len(dstFilepath) == 0 {
		dstFilepath = filepath.Base(srcFilepath)
	}
	return dstFilepath
}
