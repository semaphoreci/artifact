package storage

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	api "github.com/semaphoreci/artifact/pkg/api"
	files "github.com/semaphoreci/artifact/pkg/files"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	log "github.com/sirupsen/logrus"
)

func Push(hubClient *hub.Client, dst, src string, force bool) error {
	log.Debug("Pushing...\n")
	log.Debugf("> Source: %s\n", src)
	log.Debugf("> Destination: %s\n", dst)
	log.Debugf("> Force: %v\n", force)

	artifacts, err := locateArtifacts(src, dst)
	if err != nil {
		return err
	}

	requestType := hub.GenerateSignedURLsRequestPUSH
	if force {
		requestType = hub.GenerateSignedURLsRequestPUSHFORCE
	}

	response, err := hubClient.GenerateSignedURLs(api.RemotePaths(artifacts), requestType)
	if err != nil {
		return err
	}

	attachURLs(artifacts, response.Urls, force)
	err = doPush(force, artifacts, response.Urls)
	if err != nil {
		return err
	}

	return nil
}

func locateArtifacts(localSource, remoteDestinationPath string) ([]*api.Artifact, error) {
	isFile, err := files.IsFileSrc(localSource)
	if err != nil {
		return nil, fmt.Errorf("path '%s' does not exist locally", localSource)
	}

	if isFile {
		item := api.Artifact{RemotePath: remoteDestinationPath, LocalPath: localSource}
		return []*api.Artifact{&item}, nil
	}

	items := []*api.Artifact{}
	err = filepath.Walk(localSource, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// TODO: there's a bug here
		items = append(items, &api.Artifact{
			RemotePath: path.Join(remoteDestinationPath, filepath.Base(filename)),
			LocalPath:  filename,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

func attachURLs(items []*api.Artifact, signedURLs []*api.SignedURL, force bool) {
	i := 0
	for _, item := range items {

		/*
		 * If we are forcifully pushing artifacts,
		 * no HEAD URLs will be returned for checking the existence of an artifact.
		 * That way, each item directly relates to a single signed URL returned.
		 */
		if force {
			item.URLs = []*api.SignedURL{signedURLs[i]}
			i++
			continue
		}

		/*
		 * However, if we are not forcifully pushing artifacts,
		 * a HEAD URL + a PUT URL will be returned for each one.
		 * So in this case, each item is related to two signed URLs returned.
		 */
		item.URLs = []*api.SignedURL{signedURLs[i], signedURLs[i+1]}
		i += 2
	}
}

func doPush(force bool, artifacts []*api.Artifact, signedURLs []*api.SignedURL) error {
	client := &http.Client{}

	for _, artifact := range artifacts {
		for _, signedURL := range artifact.URLs {
			if err := signedURL.Follow(client, artifact); err != nil {
				return err
			}
		}
	}

	return nil
}
