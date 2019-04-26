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
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	ctx      = context.Background()
	bucket   *storage.BucketHandle
	locIsGCS = map[bool]string{
		true:  "Google Cloud Storage",
		false: "local file system",
	}
	isFileIsGCS = map[bool]func(string) (bool, error){
		true:  isFileGCS,
		false: isFileLFS,
	}
	isDirIsGCS = map[bool]func(string) (bool, error){
		true:  isDirGCS,
		false: isDirLFS,
	}
)

// init initializes Google Coud Storage with the given bucket name in environment variable.
// Loads credentials from environment variable too.
func init() {
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		panic(fmt.Errorf("Failed to create Google Cloud Storage client: %s", err))
	}

	bucketName := os.Getenv("BUCKET_NAME")
	bucket = client.Bucket(bucketName)
}

// isFileLFS returns if the given path points to a file in the local file system.
func isFileLFS(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		return !fi.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// isFileLFS returns if the given path points to a directory in the local file system.
func isDirLFS(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		return fi.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// isFileGCS returns if the given filename exists on the Google Cloud Storage.
func isFileGCS(filename string) (bool, error) {
	o := bucket.Object(filename)
	_, err := o.Attrs(ctx)
	if err == nil {
		return true, nil
	}
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	return false, &ErrUnknownGCS{filename}
}

// isDirGCS returns if the given directory exists on the Google Cloud Storage.
func isDirGCS(dirname string) (bool, error) {
	err := walkGCS(dirname+"/", func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return &ErrUnknownGCS{filename}
		}
		return ErrDirectoryFoundGCS
	})

	if err == ErrDirectoryFoundGCS {
		return true, nil
	}

	if _, ok := err.(*ErrUnknownGCS); ok {
		return true, err
	}

	return false, nil
}

// pushPaths returns source and destination paths to push a file to Google Cloud Storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path to the GCS Bucket.
func pushPaths(dstFilename, srcFilename string) (string, string) {
	dstFilename = utils.ToRelative(dstFilename)
	dstFilename = utils.PrefixedPathFromSource(dstFilename, srcFilename)
	return dstFilename, srcFilename
}

// doesExist returns if there is a file (or directory) at the given filename and
// location (Google Cloud Storage or local filesystem). If we found something, and there's a
// force flag set, we delete the file or directory right here.
func doesExist(dst string, force, isFile, isGCS bool) (bool, error) {
	var checker map[bool]func(string) (bool, error)
	if isFile {
		checker = isFileIsGCS
	} else {
		checker = isDirIsGCS
	}
	ok, err := checker[isGCS](dst)
	if err != nil {
		return false, err
	}
	if ok { // found, returning true
		if force {
			if isFile {
				if isGCS {
					return true, delGCS(dst)
				}
				return true, os.Remove(dst)
			}
			if isGCS {
				return true, delDirGCS(dst)
			}
			return true, os.RemoveAll(dst)
		}
		return true, &ErrAlreadyExists{dst, locIsGCS[isGCS]}
	}
	return false, nil
}

// isFileDst checks, with the given function if there is anything (file or directory)
// at the given filename and location (Google Cloud Storage or local filesystem).
func isFileDst(dst string, force, isGCS bool) error {
	ok, err := doesExist(dst, force, true, isGCS)
	if err != nil {
		return err
	}
	if ok == true { // found a file without an error
		return nil
	}
	ok, err = doesExist(dst, force, false, isGCS)
	return err
}

// isFileSrc checks, if the given source exists, and if it's a file.
func isFileSrc(src string, isGCS bool) (bool, error) {
	ok, err := isFileIsGCS[isGCS](src)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	if ok, err = isDirIsGCS[isGCS](src); err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	return false, &ErrNotFound{src, locIsGCS[isGCS]}
}

// walkGCS works like filepath.Walk, with an empty os.FileInfo, without the directories.
// Prerequisite: it shouldn't be a file.
func walkGCS(dirname string, wf filepath.WalkFunc) error {
	it := bucket.Objects(ctx, &storage.Query{Prefix: dirname})
	var emptyFi os.FileInfo
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			return nil
		}
		if err = wf(attrs.Name, emptyFi, err); err != nil {
			return err
		}
	}
	return nil
}

// pushGCS uploads a file or directory from the file system to Google Cloud Storage
// with given destination name, and human readable expire string.
func pushGCS(dst, src, expires string, force bool) error {
	expTime, err := utils.ParseRelativeAgeForHumans(expires)
	if err != nil {
		return err
	}

	isFile, err := isFileSrc(src, false)
	if err != nil {
		return err
	}

	if err = isFileDst(dst, force, true); err != nil {
		return err
	}

	if isFile {
		return pushFileGCS(dst, src, expTime)
	}

	prefLen := len(src)
	err = filepath.Walk(src, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return pushFileGCS(path.Join(dst, filename[prefLen:]), filename, expTime)
	})
	return err
}

// pushFileGCS uploads a file from the file system to Google Cloud Storage with given destination
// name and expire duration.
func pushFileGCS(dstFilename, srcFilename string, expTime time.Duration) error {
	f, err := os.Open(srcFilename)
	if err != nil {
		return fmt.Errorf("Failed to open file for pushing to Google Cloud Storage: %s", err)
	}
	defer f.Close()
	return writeGCS(dstFilename, f, expTime)
}

// pullPaths returns source and destination paths to pull a file from Google Cloud Storage.
// Source path becomes a category prefixed path to the GCS Bucket,
// destination path becomes a relative path on the file system.
func pullPaths(dstFilename, srcFilename string) (string, string) {
	srcFilename = utils.ToRelative(srcFilename)
	dstFilename = utils.PathFromSource(dstFilename, srcFilename)
	srcFilename = utils.PrefixedPath(srcFilename)
	return dstFilename, srcFilename
}

// pullGCS downloads a file or directory from the Google Cloud Storage to the file system
// with given destination and source path.
func pullGCS(dst, src string, force bool) error {
	isFile, err := isFileSrc(src, true)
	if err != nil {
		return err
	}

	if err = isFileDst(dst, force, false); err != nil {
		return err
	}

	if isFile {
		return pullFileGCS(dst, src)
	}

	prefLen := len(src)
	return walkGCS(src, func(filename string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return pullFileGCS(path.Join(dst, filename[prefLen:]), filename)
	})
}

// pullFileGCS downloads a file from the Google Cloud Storage to the file system with given source path.
func pullFileGCS(dstFilename, srcFilename string) error {
	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		return fmt.Errorf("Failed to create result dir for pulling from Google Cloud Storage: %s", err)
	}
	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		return fmt.Errorf("Failed to create result file for pulling from Google Cloud Storage: %s", err)
	}
	defer f.Close()
	return readGCS(f, srcFilename)
}

// yankPath returns path to yank a file from Google Cloud Storage.
// Path becomes a category prefixed path to the GCS Bucket.
func yankPath(filename string) string {
	filename = utils.ToRelative(filename)
	return utils.PrefixedPath(filename)
}

// yankGCS deletes a file or directory from the Google Cloud Storage.
func yankGCS(name string) error {
	isFile, err := isFileSrc(name, true)
	if err != nil {
		_, ok := err.(*ErrNotFound)
		if ok {
			return nil
		}
		return err
	}

	if isFile {
		return delGCS(name)
	}

	return delDirGCS(name)
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

// delDirGCS deletes a directory from the Google Cloud Storage with a given name.
func delDirGCS(dir string) error {
	return walkGCS(dir, func(filename string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return delGCS(filename)
	})
}

// delGCS deletes a file from the Google Cloud Storage with a given name.
func delGCS(filename string) error {
	return bucket.Object(filename).Delete(ctx)
}
