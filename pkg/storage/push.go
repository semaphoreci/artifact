package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	api "github.com/semaphoreci/artifact/pkg/api"
	files "github.com/semaphoreci/artifact/pkg/files"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	log "github.com/sirupsen/logrus"
)

type PushOptions struct {
	SourcePath          string
	DestinationOverride string
	Force               bool
}

func (o *PushOptions) RequestType() hub.GenerateSignedURLsRequestType {
	if o.Force {
		return hub.GenerateSignedURLsRequestPUSHFORCE
	}

	return hub.GenerateSignedURLsRequestPUSH
}

func Push(hubClient *hub.Client, resolver *files.PathResolver, options PushOptions) (*files.ResolvedPath, error) {
	paths, err := resolver.Resolve(files.OperationPush, options.SourcePath, options.DestinationOverride)
	if err != nil {
		return nil, err
	}

	log.Debug("Pushing...\n")
	log.Debugf("* Source: %s\n", paths.Source)
	log.Debugf("* Destination: %s\n", paths.Destination)
	log.Debugf("* Force: %v\n", options.Force)

	artifacts, err := LocateArtifacts(paths)
	if err != nil {
		return nil, err
	}

	response, err := hubClient.GenerateSignedURLs(api.RemotePaths(artifacts), options.RequestType())
	if err != nil {
		return nil, err
	}

	err = attachURLs(artifacts, response.Urls, options.Force)
	if err != nil {
		return nil, err
	}

	err = doPush(options.Force, artifacts, response.Urls)
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func LocateArtifacts(paths *files.ResolvedPath) ([]*api.Artifact, error) {
	isFile, err := files.IsFileSrc(paths.Source)
	if err != nil {
		return nil, fmt.Errorf("path '%s' does not exist locally", paths.Source)
	}

	if isFile {
		item := api.Artifact{RemotePath: paths.Destination, LocalPath: paths.Source}
		return []*api.Artifact{&item}, nil
	}

	items := []*api.Artifact{}
	err = filepath.Walk(paths.Source, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		name := filepath.ToSlash(filename)
		items = append(items, &api.Artifact{
			RemotePath: path.Join(paths.Destination, name[len(paths.Source):]),
			LocalPath:  filename,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

func attachURLs(items []*api.Artifact, signedURLs []*api.SignedURL, force bool) error {
	/*
	 * If we are forcifully pushing artifacts,
	 * no HEAD URLs will be returned for checking the existence of an artifact.
	 * That way, each item directly relates to a single signed URL returned.
	 */
	if force && len(items) != len(signedURLs) {
		return fmt.Errorf("bad number of signed URLs (%d) for forceful push - should be %d", len(signedURLs), len(items))
	}

	/*
	 * However, if we are not forcifully pushing artifacts,
	 * a HEAD URL + a PUT URL will be returned for each one.
	 * So in this case, each item is related to two signed URLs returned.
	 */
	if !force && (len(items)*2) != len(signedURLs) {
		return fmt.Errorf("bad number of signed URLs (%d) for non-forceful push - should be %d", len(signedURLs), len(items)*2)
	}

	i := 0
	for _, item := range items {
		if force {
			item.URLs = []*api.SignedURL{signedURLs[i]}
			i++
			continue
		}

		item.URLs = []*api.SignedURL{signedURLs[i], signedURLs[i+1]}
		i += 2
	}

	return nil
}

func doPush(force bool, artifacts []*api.Artifact, signedURLs []*api.SignedURL) error {
	client := retryablehttp.NewClient()
	client.RetryMax = 4
	client.RetryWaitMax = 1 * time.Second

	for _, artifact := range artifacts {
		for _, signedURL := range artifact.URLs {
			if err := signedURL.Follow(client, artifact); err != nil {
				return err
			}
		}
	}

	return nil
}
