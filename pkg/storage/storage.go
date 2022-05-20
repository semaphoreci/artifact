package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	httputil "github.com/semaphoreci/artifact/pkg/http"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	pathutil "github.com/semaphoreci/artifact/pkg/path"
	log "github.com/sirupsen/logrus"
)

// isFile returns if the given path points to a file in the local file system.
func isFile(filename string) (isF bool, ok bool) {
	fi, err := os.Stat(filename)
	if err == nil {
		return !fi.IsDir(), true
	}
	log.Error("Failed to find file '%s' to push: %v\n", filename, err)
	return false, false
}

// isDir returns if the given path points to a directory in the local file system.
func isDir(filename string) (isD bool, ok bool) {
	fi, err := os.Stat(filename)
	if err == nil {
		return fi.IsDir(), true
	}
	log.Error("Failed to find dir '%s' to push: %v\n", filename, err)
	return false, false
}

// isFileSrc checks, if the given source exists, and if it's a file.
func isFileSrc(src string) (isF bool, ok bool) {
	if isF, ok = isFile(src); !ok {
		return
	}

	if isF {
		log.Debugf("'%s' seems to be a file.\n", src)
		return
	}

	var isD bool
	if isD, ok = isDir(src); !ok {
		return
	}

	if isD {
		log.Debugf("'%s' seems to be a directory.\n", src)
		return
	}

	log.Errorf("The file or directory '%s' doesn't exist.\n", src)
	return false, false
}

// UploadFile uploads a file given by its filename to the Google Cloud Storage.
func UploadFile(u, filename string) (ok bool) {
	f, err := os.Open(filename)
	if err != nil {
		log.Error("Failed to open file '%s' for upload: %v\n", filename, err)
		return
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		log.Error("Failed to stat file '%s' for upload: %v\n", filename, err)
		return
	}

	return httputil.UploadReader(u, f, fileInfo.Size())
}

// PushPaths returns source and destination paths to push a file to storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path storage bucket.
func PushPaths(dst, src string) (string, string) {
	newDst := pathutil.ToRelative(dst)
	newDst = pathutil.PrefixedPathFromSource(newDst, src)
	newSrc := path.Clean(src)

	log.Debugln("Push parameters:")
	log.Debugf("> Input destination: '%s'\n", dst)
	log.Debugf("> Input source: '%s'\n", src)
	log.Debugf("> Output destination: '%s'\n", newDst)
	log.Debugf("> Output source: '%s'\n", newSrc)

	return newDst, newSrc
}

// Uploads a file or directory from the file system to remote storage on the
// given destination. Returns if it was a success, otherwise the error has been logged.
func Push(hubClient *hub.Client, dst, src string, force bool) (ok bool) {
	log.Debug("Pushing...\n")
	log.Debugf("> Source: %s\n", src)
	log.Debugf("> Destination: %s\n", dst)
	log.Debugf("> Force: %v\n", force)

	var isF bool
	if isF, ok = isFileSrc(src); !ok {
		return false
	}

	// local and remote paths
	var lps, rps []string

	if isF {
		rps = []string{dst}
		lps = []string{src}
	} else { // directory, getting all filenames
		rps = []string{}
		lps = []string{}
		err := filepath.Walk(src, func(filename string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			rps = append(rps, path.Join(dst, filepath.Base(filename)))
			lps = append(lps, filename)
			return nil
		})
		if err != nil {
			log.Error("Failed to walk local directory for pushing: %v\n", err)
			return false
		}
	}

	count := len(rps)
	requestType := hub.GenerateSignedURLsRequestPUSH
	if force {
		requestType = hub.GenerateSignedURLsRequestPUSHFORCE
	}

	response, err := hubClient.GenerateSignedURL(rps, requestType)
	if err != nil {
		return false
	}

	ok = ok && doPush(dst, force, count, rps, lps, response.Urls)
	if !ok {
		log.Error("File or dir not found. Please check if the source you are trying to push exists.")
		return false
	}

	return true
}

// Uploads file or directory from the file system to the remote storage.
// Returns if it was a success, otherwise the error has been logged.
func doPush(dst string, force bool, count int, rps, lps []string, signedURLs []*hub.SignedURL) (ok bool) {
	j := 0
	exist := false
	for i := 0; i < count; i, j = i+1, j+1 { // uploading files
		if !force { // needs to be checked if nothing exists there
			if exist, ok = httputil.CheckURL(signedURLs[j].URL); !ok {
				return
			}
			if exist {
				log.Errorf("The file '%s' already exists in the remote storage; delete it first, or use --force flag.\n", lps[i])
				return false
			}
			j++
		}

		log.Debugf("Uploading...\n")
		log.Debugf("> Source: %s\n", lps[i])
		log.Debugf("> Destination: %s\n", rps[i])
		if ok = UploadFile(signedURLs[j].URL, lps[i]); !ok {
			return
		}
	}

	return true
}

// PullPaths returns source and destination paths to pull a file from remote storage.
// Source path becomes a category prefixed path to the storage bucket,
// destination path becomes a relative path on the file system.
func PullPaths(dst, src string) (string, string) {
	newSrc := pathutil.ToRelative(src)
	newDst := pathutil.PathFromSource(dst, newSrc)
	newSrc = pathutil.PrefixedPath(newSrc)
	newDst = path.Clean(newDst)

	log.Debug("Paths for pulling...\n")
	log.Debugf("> Input destination: %s\n", dst)
	log.Debugf("> Input source: %s\n", src)
	log.Debugf("> Output destination: %s\n", newDst)
	log.Debugf("> Output source: %s\n", newSrc)

	return newDst, newSrc
}

// cutPrefixByDelimMulti searches for delimiter b in s from the beginning by count times.
// When the delimiter can't be found, it returns the rest string, otherwise cuts the prefix.
func cutPrefixByDelimMulti(s string, b byte, count int) string {
	for ; count > 0; count-- {
		index := strings.IndexByte(s, b)
		if index == -1 {
			return s
		}
		s = s[index+1:]
	}
	return s
}

// Downloads a file or directory from the remote storage to the file system
// with given destination and source path.
func Pull(hubClient *hub.Client, dst, src string, force bool) (ok bool) {
	log.Debug("Pulling...\n")
	log.Debugf("> Source: %s\n", src)
	log.Debugf("> Destination: %s\n", dst)
	log.Debugf("> Force: %v\n", force)

	ps := []string{src}

	response, err := hubClient.GenerateSignedURL(ps, hub.GenerateSignedURLsRequestPULL)
	if err != nil {
		return false
	}

	ok = doPull(dst, src, force, response.Urls)

	if !ok {
		log.Error("Artifact not found. Please check if the artifact you are trying to pull exists.")
		return false
	}

	return true
}

func doPull(dst, src string, force bool, URLs []*hub.SignedURL) (ok bool) {
	if len(URLs) == 1 { // one file only
		signedURL := URLs[0]
		obj, err := signedURL.GetObject()
		if err != nil {
			log.Errorf("Error finding object to pull from URL '%s': %v\n", signedURL.URL, err)
			return
		}

		// removing <project-name>/<category>/<projectID>/ prefix
		obj = cutPrefixByDelimMulti(obj, '/', 3)

		// TODO: pretty sure this is never true
		if obj == src { // they are the same: requested single file pull
			return PullFile(dst, signedURL.URL, force)
		} // otherwise it will be downloaded as a directory
	}

	prefLen := len(src)
	for _, signedURL := range URLs { // iterate all urls and put them in a directory structure
		obj, err := signedURL.GetObject()
		if err != nil {
			log.Errorf("Error finding object to pull from URL '%s': %v\n", signedURL.URL, err)
			return
		}

		if ok = PullFile(path.Join(dst, obj[prefLen:]), signedURL.URL, force); !ok {
			return
		}
	}

	return true
}

func PullFile(dstFilename, u string, force bool) (ok bool) {
	log.Debug("Downloading...\n")
	log.Debugf("> URL: %s\n", u)
	log.Debugf("> Destination: %s\n", dstFilename)

	if !force {
		if _, err := os.Stat(dstFilename); err == nil {
			log.Errorf("The file '%s' already exists locally; delete it first, or use --force flag.\n", dstFilename)
			return
		}
	}

	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		log.Errorf("Failed to create dir for pulling: %v\n", err)
		return
	}

	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		log.Error("Failed to create file for pulling: %v\n", err)
		return
	}

	defer f.Close()
	ok = httputil.DownloadWriter(u, f)

	log.Debugf("Pull file result: %v\n", ok)
	return ok
}

// Returns path to yank a file from the remote storage.
// Path becomes a category prefixed path to the storage bucket.
func YankPath(f string) string {
	newF := pathutil.ToRelative(f)
	newF = pathutil.PrefixedPath(newF)

	// TODO: should we be logging this at all?
	log.Debugf("> Input file: %s\n", f)
	log.Debugf("> Output file: %s\n", newF)

	return newF
}

// Deletes a file or directory from the remote storage
func Yank(hubClient *hub.Client, name string) error {
	response, err := hubClient.GenerateSignedURL([]string{name}, hub.GenerateSignedURLsRequestYANK)
	if err != nil {
		return err
	}

	err = doYank(response.Urls)
	if err != nil {
		log.Error("Error deleting artifact. Make sure the artifact you are trying to yank exists.\n")
		return err
	}

	return nil
}

func doYank(URLs []*hub.SignedURL) error {
	log.Debug("Deleting...\n")
	for _, u := range URLs {
		log.Debugf("> DELETE '%s'...\n", u.URL)
		if ok := httputil.DeleteURL(u.URL); !ok {
			return fmt.Errorf("error deleting artifact")
		}
	}

	return nil
}
