package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	files "github.com/semaphoreci/artifact/pkg/files"
	httputil "github.com/semaphoreci/artifact/pkg/http"
	hub "github.com/semaphoreci/artifact/pkg/hub"
	log "github.com/sirupsen/logrus"
)

func UploadFile(u, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file '%s' for upload: %v", filename, err)
	}

	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file '%s' for upload: %v", filename, err)
	}

	return httputil.UploadReader(u, f, fileInfo.Size())
}

// Uploads a file or directory from the file system to remote storage on the
// given destination. Returns if it was a success, otherwise the error has been logged.
func Push(hubClient *hub.Client, dst, src string, force bool) error {
	log.Debug("Pushing...\n")
	log.Debugf("> Source: %s\n", src)
	log.Debugf("> Destination: %s\n", dst)
	log.Debugf("> Force: %v\n", force)

	isF, err := files.IsFileSrc(src)
	if err != nil {
		return fmt.Errorf("path '%s' does not exist locally", src)
	}

	// TODO: refactor this code as well
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
			return fmt.Errorf("failed walking '%s': %v", src, err)
		}
	}

	count := len(rps)
	requestType := hub.GenerateSignedURLsRequestPUSH
	if force {
		requestType = hub.GenerateSignedURLsRequestPUSHFORCE
	}

	response, err := hubClient.GenerateSignedURL(rps, requestType)
	if err != nil {
		return err
	}

	err = doPush(dst, force, count, rps, lps, response.Urls)
	if err != nil {
		log.Errorf("Error pushing artifact: %v\n", err)
		return err
	}

	return nil
}

// Uploads file or directory from the file system to the remote storage.
// Returns if it was a success, otherwise the error has been logged.
func doPush(dst string, force bool, count int, rps, lps []string, signedURLs []*hub.SignedURL) error {
	j := 0

	// TODO: refactor
	for i := 0; i < count; i, j = i+1, j+1 {
		if !force {
			exist, err := httputil.Exists(signedURLs[j].URL)
			if err != nil {
				return err
			}

			if exist {
				return fmt.Errorf("the file '%s' already exists in the remote storage; delete it first, or use --force flag", lps[i])
			}

			j++
		}

		log.Debugf("Uploading...\n")
		log.Debugf("> Source: %s\n", lps[i])
		log.Debugf("> Destination: %s\n", rps[i])

		// TODO: why do we need a loop here at all, if we are returning at the tail of it?
		return UploadFile(signedURLs[j].URL, lps[i])
	}

	return nil
}

// Downloads a file or directory from the remote storage to the file system
// with given destination and source path.
func Pull(hubClient *hub.Client, dst, src string, force bool) error {
	log.Debug("Pulling...\n")
	log.Debugf("> Source: %s\n", src)
	log.Debugf("> Destination: %s\n", dst)
	log.Debugf("> Force: %v\n", force)

	ps := []string{src}

	response, err := hubClient.GenerateSignedURL(ps, hub.GenerateSignedURLsRequestPULL)
	if err != nil {
		return err
	}

	return doPull(dst, src, force, response.Urls)
}

func doPull(dst, src string, force bool, URLs []*hub.SignedURL) error {
	prefLen := len(src)
	for _, signedURL := range URLs { // iterate all urls and put them in a directory structure
		obj, err := signedURL.GetObject()
		if err != nil {
			return err
		}

		err = PullFile(path.Join(dst, obj[prefLen:]), signedURL.URL, force)
		if err != nil {
			return err
		}
	}

	return nil
}

func PullFile(dstFilename, u string, force bool) error {
	log.Debug("Downloading...\n")
	log.Debugf("> URL: %s\n", u)
	log.Debugf("> Destination: %s\n", dstFilename)

	if !force {
		if _, err := os.Stat(dstFilename); err == nil {
			return fmt.Errorf("%s already exists locally; delete it first, or use --force flag", dstFilename)
		}
	}

	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	defer f.Close()

	return httputil.DownloadWriter(u, f)
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
