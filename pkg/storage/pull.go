package storage

import (
	"fmt"
	"os"
	"path"

	api "github.com/semaphoreci/artifact/pkg/api"
	"github.com/semaphoreci/artifact/pkg/files"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	log "github.com/sirupsen/logrus"
)

type PullOptions struct {
	SourcePath          string
	DestinationOverride string
	Force               bool
}

type PullStats struct {
	FileCount int
	TotalSize int64
}

func Pull(hubClient *hub.Client, resolver *files.PathResolver, options PullOptions) (*files.ResolvedPath, *PullStats, error) {
	paths, err := resolver.Resolve(files.OperationPull, options.SourcePath, options.DestinationOverride)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("Pulling...\n")
	log.Debugf("* Source: %s\n", paths.Source)
	log.Debugf("* Destination: %s\n", paths.Destination)
	log.Debugf("* Force: %v\n", options.Force)

	response, err := hubClient.GenerateSignedURLs([]string{paths.Source}, hub.GenerateSignedURLsRequestPULL)
	if err != nil {
		return nil, nil, err
	}

	artifacts, err := buildArtifacts(response.Urls, paths, options.Force)
	if err != nil {
		return nil, nil, err
	}

	stats, err := doPull(artifacts)
	if err != nil {
		return nil, nil, err
	}

	return paths, stats, nil
}

func buildArtifacts(signedURLs []*api.SignedURL, paths *files.ResolvedPath, force bool) ([]*api.Artifact, error) {
	artifacts := []*api.Artifact{}

	for _, signedURL := range signedURLs {
		obj, err := signedURL.GetObject()
		if err != nil {
			return nil, err
		}

		localPath := path.Join(paths.Destination, obj[len(paths.Source):])

		if !force {
			if _, err := os.Stat(localPath); err == nil {
				return nil, fmt.Errorf("'%s' already exists locally; delete it first, or use --force flag", localPath)
			}
		}

		artifacts = append(artifacts, &api.Artifact{
			RemotePath: obj,
			LocalPath:  localPath,
			URLs:       []*api.SignedURL{signedURL},
		})
	}

	return artifacts, nil
}

func doPull(artifacts []*api.Artifact) (*PullStats, error) {
	client := newHTTPClient()
	stats := &PullStats{}

	for _, artifact := range artifacts {
		for _, signedURL := range artifact.URLs {
			if err := signedURL.Follow(client, artifact); err != nil {
				return nil, err
			}

			// Get file size after successful download
			if fileInfo, err := os.Stat(artifact.LocalPath); err == nil {
				stats.FileCount++
				stats.TotalSize += fileInfo.Size()
			}
		}
	}

	return stats, nil
}
