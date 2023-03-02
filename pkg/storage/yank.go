package storage

import (
	api "github.com/semaphoreci/artifact/pkg/api"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	log "github.com/sirupsen/logrus"
)

// Deletes a file or directory from the remote storage
func Yank(hubClient *hub.Client, name string, verbose bool) error {
	response, err := hubClient.GenerateSignedURLs([]string{name}, hub.GenerateSignedURLsRequestYANK, verbose)
	if err != nil {
		return err
	}

	err = doYank(response.Urls)
	if err != nil {
		log.Errorf("Error deleting artifact. Make sure the artifact you are trying to yank exists: %v\n", err)
		return err
	}

	return nil
}

func doYank(URLs []*api.SignedURL) error {
	client := newHTTPClient()

	for _, u := range URLs {
		// The hub is not returning the method for yank operations, so we fill it here
		u.Method = "DELETE"
		if err := u.Follow(client, nil); err != nil {
			return err
		}
	}

	return nil
}
