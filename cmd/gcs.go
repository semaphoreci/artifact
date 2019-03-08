package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/semaphoreci/artifact/cmd/utils"
	"google.golang.org/api/option"
)

var (
	ctx    = context.Background()
	bucket *storage.BucketHandle
)

// initGCS initializes Google Coud Storage with the given bucket name.
// Loads credentials from environment variable.
func initGCS() error {
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		return fmt.Errorf("Failed to create Google Cloud Storage client: %s", err)
	}

	bucketName := os.Getenv("BUCKET_NAME")
	bucket = client.Bucket(bucketName)
	return nil
}

// pushFileGCS uploads a file from the file system to Google Cloud Storage with given destination
// path and name, and human readable expire string.
func pushFileGCS(dstDir, dstFilename, srcFilename, expires string) error {
	expTime, err := utils.ParseRelativeAgeForHumans(expires)
	if err != nil {
		return err
	}
	if len(dstFilename) == 0 {
		dstFilename = filepath.Base(srcFilename)
	}
	var f *os.File
	if f, err = os.Open(srcFilename); err != nil {
		return err
	}
	defer f.Close()
	return writeGCS(path.Join(dstDir, dstFilename), f, expTime)
}

// pullFileGCS downloads a file from the Google Cloud Storage to the file system with given source
// path and name.
func pullFileGCS(dstDir, dstFilename, srcFilename string) (err error) {
	if len(dstFilename) == 0 {
		dstFilename = filepath.Base(srcFilename)
	}
	var f *os.File
	if f, err = os.Create(srcFilename); err != nil {
		return
	}
	defer f.Close()
	return readGCS(f, path.Join(dstDir, dstFilename))
}

func yankFileGCS(dir, filename string) error {
	// TODO
	return nil
}

// writeGCS uploads a file from an io Reader to Google Cloud Storage with given destination
// path and name, and an expiration duration.
func writeGCS(dstFilename string, srcReader io.Reader, expires time.Duration) error {
	w := bucket.Object(dstFilename).NewWriter(ctx)
	defer w.Close()
	_, err := io.Copy(w, srcReader)
	// TODO: set expire
	return err
}

// readGCS downloads a file to an io Writer from Google Cloud Storage with given source
// path and name.
func readGCS(dstWriter io.Writer, srcFilename string) error {
	r, err := bucket.Object(srcFilename).NewReader(ctx)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstWriter, r)
	return err
}

// delGCS deletes a file from the Google Cloud Storage with a given name.
func delGCS(filename string) error {
	return bucket.Object(filename).Delete(ctx)
}
