package storage

import (
	"fmt"
	"net/http"
	"os"
	"path"

	api "github.com/semaphoreci/artifact/pkg/api"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	log "github.com/sirupsen/logrus"
)

func Pull(hubClient *hub.Client, dst, src string, force bool) error {
	log.Debug("Pulling...\n")
	log.Debugf("* Source: %s\n", src)
	log.Debugf("* Destination: %s\n", dst)
	log.Debugf("* Force: %v\n", force)

	response, err := hubClient.GenerateSignedURLs([]string{src}, hub.GenerateSignedURLsRequestPULL)
	if err != nil {
		return err
	}

	artifacts, err := buildArtifacts(response.Urls, dst, src, force)
	if err != nil {
		return err
	}

	return doPull(force, artifacts, response.Urls)
}

func buildArtifacts(signedURLs []*api.SignedURL, localPath, remotePath string, force bool) ([]*api.Artifact, error) {
	artifacts := []*api.Artifact{}

	for _, signedURL := range signedURLs {
		obj, err := signedURL.GetObject()
		if err != nil {
			return nil, err
		}

		// TODO: figure out if there's a better way to find this localPath
		localPath := path.Join(localPath, obj[len(remotePath):])

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

func doPull(force bool, artifacts []*api.Artifact, signedURLs []*api.SignedURL) error {
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
