package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/semaphoreci/artifact/cmd/utils"
	"google.golang.org/api/option"
)

var (
	ctx        = context.Background()
	bucketName string
	bucket     *storage.BucketHandle
	gcsConf    GcsConfig
)

// GcsConfig contains config options related to Google Cloud Storage.
type GcsConfig struct {
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
}

// initGCS initializes Google Coud Storage with the given bucket name.
// Loads credentials from environment variable.
func initGCS(bName string) error {
	bucketName = bName
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	f, err := os.Open(credFile)
	if err != nil {
		return fmt.Errorf("Failed to open Google Cloud Storage credentials file: %s", err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	if err = d.Decode(&gcsConf); err != nil {
		return fmt.Errorf("Failed to decode Google Cloud Storage credentials file: %s", err)
	}

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		return fmt.Errorf("Failed to create Google Cloud Storage client: %s", err)
	}

	bucket = client.Bucket(bucketName)
	return nil
}

// signedurlGCS creates a signed, expirable url to a Google Cloud Storage Object.
func signedurlGCS(filename, method string, expires time.Time) (url string, err error) {
	if url, err = storage.SignedURL(bucketName, filename, &storage.SignedURLOptions{
		GoogleAccessID: gcsConf.ClientEmail,
		PrivateKey:     []byte(gcsConf.PrivateKey),
		Method:         method,
		Expires:        expires,
	}); err != nil {
		err = fmt.Errorf("Failed to create Google Cloud Storage signed url: %s", err)
	}
	return
}

// writeGCS uploads a file from the file system to Google Cloud Storage with given destination
// path and name, and human readable expire string.
func writeGCS(dstDir, dstFilename, srcFilename, expires string) error {
	expTime, err := utils.ParseRelativeAgeForHumans
	if err != nil {
		return err
	}
	if expTime == -1 {
		// TODO: create non-expire object, and upload file
	} else {
		// TODO: create expire signed url, and upload file
	}
}
