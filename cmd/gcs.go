package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
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

// pushPaths returns source and destination paths to push a file to Google Cloud Storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path to the GCS Bucket.
func pushPaths(category, dstFilename, srcFilename string) (dstPath, srcPath string) {
	srcFilename = utils.ToRelative(srcFilename)
	dstFilename = utils.PrefixedPathFromSource(category, dstFilename, srcFilename)
	return dstFilename, srcFilename
}

// pushFileGCS uploads a file from the file system to Google Cloud Storage with given category,
// destination name, and human readable expire string.
func pushFileGCS(category, dstFilename, srcFilename, expires string) (string, string, error) {
	expTime, err := utils.ParseRelativeAgeForHumans(expires)
	if err != nil {
		return "", "", err
	}
	dstFilename, srcFilename = pushPaths(category, dstFilename, srcFilename)
	var f *os.File
	if f, err = os.Open(srcFilename); err != nil {
		return "", "", fmt.Errorf("Failed to open file for pushing to Google Cloud Storage: %s", err)
	}
	defer f.Close()
	return dstFilename, srcFilename, writeGCS(dstFilename, f, expTime)
}

// pullPaths returns source and destination paths to pull a file from Google Cloud Storage.
// Source path becomes a category prefixed path to the GCS Bucket,
// destination path becomes a relative path on the file system.
func pullPaths(category, dstFilename, srcFilename string) (dstPath, srcPath string) {
	dstFilename = utils.ToRelative(utils.PathFromSource(dstFilename, srcFilename))
	srcFilename = utils.PrefixedPath(category, srcFilename)
	return dstFilename, srcFilename
}

// pullFileGCS downloads a file from the Google Cloud Storage to the file system with given category,
// and source path.
func pullFileGCS(category, dstFilename, srcFilename string) (string, string, error) {
	dstFilename, srcFilename = pullPaths(category, dstFilename, srcFilename)
	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		return "", "", fmt.Errorf("Failed to create result dir for pulling from Google Cloud Storage: %s", err)
	}
	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		return "", "", fmt.Errorf("Failed to create result file for pulling from Google Cloud Storage: %s", err)
	}
	defer f.Close()
	return dstFilename, srcFilename, readGCS(f, srcFilename)
}

// yankFileGCS deletes a file from the Google Cloud Storage with given category and filename.
func yankFileGCS(category, filename string) (string, error) {
	filename = utils.PrefixedPath(category, filename)
	return filename, delGCS(filename)
}

// writeGCS uploads a file from an io Reader to Google Cloud Storage with given destination
// path and name, and an expiration duration.
func writeGCS(dstFilename string, srcReader io.Reader, expires time.Duration) error {
	o := bucket.Object(dstFilename)
	_, err := o.Attrs(ctx)
	if err != storage.ErrObjectNotExist {
		return fmt.Errorf("The file '%s' already exists in the Google Cloud Storage", dstFilename)
	}
	w := o.NewWriter(ctx)
	defer w.Close()
	_, err = io.Copy(w, srcReader)
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
