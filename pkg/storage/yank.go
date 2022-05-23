package storage

import (
	api "github.com/semaphoreci/artifact/pkg/api"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	log "github.com/sirupsen/logrus"
)

// Deletes a file or directory from the remote storage
func Yank(hubClient *hub.Client, name string) error {
	response, err := hubClient.GenerateSignedURLs([]string{name}, hub.GenerateSignedURLsRequestYANK)
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
	for _, u := range URLs {
		// TODO: there's an issue with artifacthub not returning the method for yank operations
		u.Method = "DELETE"
		if err := u.Follow(nil); err != nil {
			return err
		}
	}

	return nil
}
