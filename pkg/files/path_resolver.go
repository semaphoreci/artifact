package files

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

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
	source = filepath.ToSlash(source)
	destination = filepath.ToSlash(destination)

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
	localDestination := path.Clean(pathFromSource(destination, remoteSource))
	remoteSource = r.PrefixedPath(remoteSource)

	log.Debug("Resolved paths.\n")
	log.Debugf("* Local destination: %s\n", localDestination)
	log.Debugf("* Remote source: %s\n", remoteSource)

	return &ResolvedPath{Source: remoteSource, Destination: localDestination}
}

func (r *PathResolver) Push(source, destination string) *ResolvedPath {
	remoteDestination := ToRelative(destination)
	remoteDestination = r.PrefixedPath(pathFromSource(remoteDestination, source))
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
	prefixedFile := r.PrefixedPath(ToRelative(file))

	log.Debug("Resolved paths.\n")
	log.Debugf("* Remote file: %s\n", prefixedFile)

	return &ResolvedPath{Source: prefixedFile}
}

/*
 * Get resource-prefixed paths for paths in remote storage.
 *
 * For project: artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip
 * For workflow: artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip
 * For job: artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip
 */
func (r *PathResolver) PrefixedPath(filepath string) string {
	return path.Join("artifacts", r.ResourceTypePlural, r.ResourceIdentifier, filepath)
}

// If no destination is set, we take the destination path from the source.
func pathFromSource(destination, source string) string {
	if destination == "" {
		return path.Base(source)
	}

	return destination
}
