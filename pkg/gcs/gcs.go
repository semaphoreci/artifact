package gcs

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/semaphoreci/artifact/internal"
	"github.com/semaphoreci/artifact/pkg/utils"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	googleCredFile = "~/.artifact/credentials.json"
	randPostfixLen = 6
	randChars      = "abcdefghijklmnopqrstuvwxyz0123456789"
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
	h, err := homedir.Dir()
	if err != nil {
		panic(fmt.Errorf("Failed to find home directory: %s", err))
	}
	credFile := strings.Replace(googleCredFile, "~", h, 1)

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		panic(fmt.Errorf("Failed to create Google Cloud Storage client: %s", err))
	}

	bucketName := os.Getenv("SEMAPHORE_ARTIFACT_BUCKET_NAME")
	fmt.Println("artifact inited with bucket name", bucketName)
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
	return false, &internal.ErrUnknownGCS{filename}
}

// isDirGCS returns if the given directory exists on the Google Cloud Storage.
func isDirGCS(dirname string) (bool, error) {
	err := walkGCS(dirname+"/", func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return &internal.ErrUnknownGCS{filename}
		}
		return internal.ErrDirectoryFoundGCS
	})

	if err == internal.ErrDirectoryFoundGCS {
		return true, nil
	}

	if _, ok := err.(*internal.ErrUnknownGCS); ok {
		return true, err
	}

	return false, nil
}

// PushPaths returns source and destination paths to push a file to Google Cloud Storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path to the GCS Bucket.
func PushPaths(dstFilename, srcFilename string) (string, string) {
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
					return true, DelGCS(dst)
				}
				return true, os.Remove(dst)
			}
			if isGCS {
				return true, DelDirGCS(dst)
			}
			return true, os.RemoveAll(dst)
		}
		return true, &internal.ErrAlreadyExists{dst, locIsGCS[isGCS]}
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
	return false, &internal.ErrNotFound{src, locIsGCS[isGCS]}
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

// PushGCS uploads a file or directory from the file system to Google Cloud Storage to given destination
// with a human readable expire string. Returns expire filename and error, if happened any.
func PushGCS(dst, src, expires string, force bool) (string, error) {
	expTime, err := utils.ParseRelativeAgeForHumans(expires)
	if err != nil {
		return "", err
	}

	isFile, err := isFileSrc(src, false)
	if err != nil {
		return "", err
	}

	if err = isFileDst(dst, force, true); err != nil {
		return "", err
	}

	if isFile {
		return PushFileGCS(dst, src, expTime)
	}

	prefLen := len(src)
	anySuccess := false
	err = filepath.Walk(src, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if _, err = PushFileGCS(path.Join(dst, filename[prefLen:]), filename, 0); err == nil {
			anySuccess = true
		}
		return err
	})

	if anySuccess { // there was at least one object created: expire needs to be created also
		return CreateExpireFile(dst, expTime), err
	}

	return "", err
}

func randomString() (string, error) {
	output := make([]byte, randPostfixLen)
	randomness := make([]byte, randPostfixLen)

	// generate some random bytes, this shouldn't fail
	_, err := rand.Read(randomness)
	if err != nil {
		return "", fmt.Errorf("Random number generation failed: %s", err)
	}

	// fill output
	l := uint8(len(randChars))
	for pos := 0; pos < randPostfixLen; pos++ {
		random := uint8(randomness[pos])   // get random item
		randomPos := random % uint8(l)     // random % length
		output[pos] = randChars[randomPos] // put into output
	}
	return string(output), nil
}

// PushFileGCS uploads a file from the file system to Google Cloud Storage with given destination
// name and expire duration. Returns expire filename or error, if happened any.
func PushFileGCS(dstFilename, srcFilename string, expTime time.Duration) (string, error) {
	f, err := os.Open(srcFilename)
	if err != nil {
		return "", fmt.Errorf("Failed to open file for pushing to Google Cloud Storage: %s", err)
	}
	defer f.Close()

	if err = WriteGCS(dstFilename, f); err != nil {
		return "", fmt.Errorf("Failed to write object on Google Cloud Storage: %s", err)
	}

	if expTime > 0 {
		return CreateExpireFile(dstFilename, expTime), nil
	}
	return "", nil
}

// CreateExpireFile creates an expire descriptor file to the Google Cloud Storage.
func CreateExpireFile(dst string, expTime time.Duration) string {
	if expTime < 1 {
		return ""
	}

	randPostfix, err := randomString()
	if err != nil {
		// TODO: some logging would be nice here, to let the system know, that something is down
		return "" // this is not a critical service
	}
	expFilename := strconv.FormatInt(time.Now().Add(expTime).Unix(), 10)
	var b strings.Builder
	expFilename = path.Join(utils.ExpirePrefix, expFilename)
	b.WriteString(expFilename)
	b.WriteByte('-')
	b.WriteString(randPostfix)
	expFilename = b.String()
	if err = WriteGCS(expFilename, strings.NewReader(dst)); err != nil {
		// TODO: some logging would be nice here, to let the system know, that something is down
		return "" // this is not a critical service
	}
	return expFilename
}

// PullPaths returns source and destination paths to pull a file from Google Cloud Storage.
// Source path becomes a category prefixed path to the GCS Bucket,
// destination path becomes a relative path on the file system.
func PullPaths(dstFilename, srcFilename string) (string, string) {
	srcFilename = utils.ToRelative(srcFilename)
	dstFilename = utils.PathFromSource(dstFilename, srcFilename)
	srcFilename = utils.PrefixedPath(srcFilename)
	return dstFilename, srcFilename
}

// PullGCS downloads a file or directory from the Google Cloud Storage to the file system
// with given destination and source path.
func PullGCS(dst, src string, force bool) error {
	isFile, err := isFileSrc(src, true)
	if err != nil {
		return err
	}

	if err = isFileDst(dst, force, false); err != nil {
		return err
	}

	if isFile {
		return PullFileGCS(dst, src)
	}

	prefLen := len(src)
	return walkGCS(src, func(filename string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return PullFileGCS(path.Join(dst, filename[prefLen:]), filename)
	})
}

// PullFileGCS downloads a file from the Google Cloud Storage to the file system with given source path.
func PullFileGCS(dstFilename, srcFilename string) error {
	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		return fmt.Errorf("Failed to create result dir for pulling from Google Cloud Storage: %s", err)
	}
	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		return fmt.Errorf("Failed to create result file for pulling from Google Cloud Storage: %s", err)
	}
	defer f.Close()
	return ReadGCS(f, srcFilename)
}

// YankPath returns path to yank a file from Google Cloud Storage.
// Path becomes a category prefixed path to the GCS Bucket.
func YankPath(filename string) string {
	filename = utils.ToRelative(filename)
	return utils.PrefixedPath(filename)
}

// YankGCS deletes a file or directory from the Google Cloud Storage.
func YankGCS(name string) error {
	isFile, err := isFileSrc(name, true)
	if err != nil {
		_, ok := err.(*internal.ErrNotFound)
		if ok {
			return nil
		}
		return err
	}

	if isFile {
		return DelGCS(name)
	}

	return DelDirGCS(name)
}

// WriteGCS uploads a file from an io Reader to Google Cloud Storage with given destination
// path and name.
func WriteGCS(dstFilename string, srcReader io.Reader) error {
	w := bucket.Object(dstFilename).NewWriter(ctx)
	defer w.Close()
	_, err := io.Copy(w, srcReader)
	return err
}

// ReadGCS downloads a file to an io Writer from Google Cloud Storage with given source
// path and name.
func ReadGCS(dstWriter io.Writer, srcFilename string) error {
	r, err := bucket.Object(srcFilename).NewReader(ctx)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstWriter, r)
	return err
}

// DelDirGCS deletes a directory from the Google Cloud Storage with a given name.
func DelDirGCS(dir string) error {
	return walkGCS(dir, func(filename string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return DelGCS(filename)
	})
}

// DelGCS deletes a file from the Google Cloud Storage with a given name.
func DelGCS(filename string) error {
	return bucket.Object(filename).Delete(ctx)
}
