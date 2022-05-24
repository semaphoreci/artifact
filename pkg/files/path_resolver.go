package files

import (
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	ResourceTypeProject  = "project"
	ResourceTypeWorkflow = "workflow"
	ResourceTypeJob      = "job"
	OperationPush        = "push"
	OperationPull        = "pull"
	OperationYank        = "yank"
)

type PathResolver struct {
	ResourceType       string
	ResourceTypePlural string
	ResourceIdentifier string
}

func NewPathResolver(resourceType, resourceId string) (*PathResolver, error) {
	switch resourceType {
	case ResourceTypeProject:
		id := id(os.Getenv("SEMAPHORE_PROJECT_ID"), resourceId)
		if id == "" {
			return nil, fmt.Errorf("project ID is not set. Please use the SEMAPHORE_PROJECT_ID environment variable or the --project-id parameter to configure it")
		}

		return &PathResolver{
			ResourceType:       resourceType,
			ResourceTypePlural: "projects",
			ResourceIdentifier: id,
		}, nil
	case ResourceTypeWorkflow:
		id := id(os.Getenv("SEMAPHORE_WORKFLOW_ID"), resourceId)
		if id == "" {
			return nil, fmt.Errorf("workflow ID is not set. Please use the SEMAPHORE_WORKFLOW_ID environment variable or the --workflow-id parameter to configure it")
		}

		return &PathResolver{
			ResourceType:       resourceType,
			ResourceTypePlural: "workflows",
			ResourceIdentifier: id,
		}, nil
	case ResourceTypeJob:
		id := id(os.Getenv("SEMAPHORE_JOB_ID"), resourceId)
		if id == "" {
			return nil, fmt.Errorf("project ID is not set. Please use the SEMAPHORE_JOB_ID environment variable or the --job-id parameter to configure it")
		}

		return &PathResolver{
			ResourceType:       resourceType,
			ResourceTypePlural: "jobs",
			ResourceIdentifier: id,
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized resource type '%s'", resourceType)
	}
}

func id(defaultValue, override string) string {
	if override == "" {
		return defaultValue
	}

	return override
}

type ResolvedPath struct {
	Source      string
	Destination string
}

func (r *PathResolver) Resolve(operation, source, destination string) (*ResolvedPath, error) {
	switch operation {
	case OperationPush:
		return r.Push(source, destination), nil
	case OperationPull:
		return r.Pull(source, destination), nil
	case OperationYank:
		return r.Yank(source), nil
	default:
		return nil, fmt.Errorf("unrecognized operation '%s'", operation)
	}
}

func (r *PathResolver) Pull(source, destination string) *ResolvedPath {
	remoteSource := ToRelative(source)
	localDestination := path.Clean(PathFromSource(destination, remoteSource))
	remoteSource = r.prefixedPath(remoteSource)

	log.Debug("Resolved paths.\n")
	log.Debugf("* Local destination: %s\n", localDestination)
	log.Debugf("* Remote source: %s\n", remoteSource)

	return &ResolvedPath{Source: remoteSource, Destination: localDestination}
}

func (r *PathResolver) Push(source, destination string) *ResolvedPath {
	remoteDestination := r.prefixedPathFromSource(ToRelative(destination), source)
	localSource := path.Clean(source)

	log.Debug("Resolved paths.\n")
	log.Debugf("* Remote destination: '%s'\n", remoteDestination)
	log.Debugf("* Local source: '%s'\n", localSource)

	return &ResolvedPath{
		Source:      localSource,
		Destination: remoteDestination,
	}
}

func (r *PathResolver) Yank(file string) *ResolvedPath {
	prefixedFile := r.prefixedPath(ToRelative(file))

	log.Debug("Resolved paths.\n")
	log.Debugf("* Remote file: %s\n", prefixedFile)

	return &ResolvedPath{Source: prefixedFile}
}

// PrefixedPath returns paths for Google Cloud Storage.
// For project files, it returns like: artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip
// For workflow files, it returns like: artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip
// For job files, it returns like: artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip
func (r *PathResolver) prefixedPath(filepath string) string {
	return path.Join("artifacts", r.ResourceTypePlural, r.ResourceIdentifier, filepath)
}

// PrefixedPathFromSource returns a path for Google Cloud Storage, where destination filename can be
// empty. In this case filename is gained from source filename, eg. uploading /from/this/path/x.zip
// with empty --destination to the project will return artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip,
// but with --destination=y.zip will result in artifacts/projects/<SEMAPHORE_PROJECT_ID>/y.zip .
func (r *PathResolver) prefixedPathFromSource(dstFilepath, srcFilepath string) string {
	dstFilepath = PathFromSource(dstFilepath, srcFilepath)
	return r.prefixedPath(dstFilepath)
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
