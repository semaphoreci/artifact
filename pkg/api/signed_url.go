package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/semaphoreci/artifact/pkg/common"
	log "github.com/sirupsen/logrus"
)

// SignedURL contains an url and its method type.
type SignedURL struct {
	URL    string `json:"url,omitempty"`
	Method string `json:"method,omitempty"`
}

func (u *SignedURL) Follow(client *http.Client, artifact *Artifact) error {
	switch u.Method {
	case "HEAD":
		return u.head(client, artifact)

	case "GET":
		return u.get(client, artifact)

	case "PUT":
		return u.put(client, artifact)

	case "DELETE":
		return u.delete(client, artifact)

	default:
		return fmt.Errorf("method '%s' not implemented", u.Method)
	}
}

func (u *SignedURL) head(client *http.Client, artifact *Artifact) error {
	log.Debugf("HEAD '%s'...\n", u.URL)

	resp, err := client.Head(u.URL)
	if err != nil {
		return fmt.Errorf("error executing HEAD '%s': %v", u, err)
	}

	defer resp.Body.Close()

	log.Debugf("HEAD request got %d response.\n", resp.StatusCode)
	if common.IsStatusOK(resp.StatusCode) {
		return fmt.Errorf("'%s' already exists in the remote storage; delete it first, or use --force flag", artifact.LocalPath)
	}

	return nil
}

func (u *SignedURL) put(client *http.Client, artifact *Artifact) error {
	log.Debugf("Opening '%s' for upload...\n", artifact.LocalPath)

	f, err := os.Open(artifact.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %v", artifact.LocalPath, err)
	}

	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat '%s': %v", artifact.LocalPath, err)
	}

	var contentBody io.Reader = f

	// If the file has no bytes, we need to use http.NoBody
	// See https://cs.opensource.google/go/go/+/refs/tags/go1.18.2:src/net/http/request.go;l=920
	if fileInfo.Size() == 0 {
		log.Debugf("'%s' is empty.", artifact.LocalPath)
		contentBody = http.NoBody
	}

	log.Debugf("PUT '%s'...\n", u.URL)
	req, err := http.NewRequest("PUT", u.URL, contentBody)
	if err != nil {
		return fmt.Errorf("failed to create new http request: %v", err)
	}

	req.ContentLength = fileInfo.Size()
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %v", err)
	}

	defer response.Body.Close()

	log.Debugf("PUT request got %d response.\n", response.StatusCode)
	if !common.IsStatusOK(response.StatusCode) {
		return fmt.Errorf("request failed with %d", response.StatusCode)
	}

	return nil
}

func (u *SignedURL) get(client *http.Client, artifact *Artifact) error {
	log.Debugf("GET '%s'...\n", u.URL)

	parentDir := filepath.Dir(artifact.LocalPath)
	err := os.MkdirAll(parentDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create parent directory '%s': %v", parentDir, err)
	}

	var f *os.File
	if f, err = os.Create(artifact.LocalPath); err != nil {
		return fmt.Errorf("failed to create local file '%s': %v", artifact.LocalPath, err)
	}

	defer f.Close()

	req, err := http.NewRequest("GET", u.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create GET request: %v", err)
	}

	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute GET request: %v", err)
	}

	log.Debugf("GET request got %d response.\n", response.StatusCode)
	if !common.IsStatusOK(response.StatusCode) {
		return fmt.Errorf("GET failed with %d", response.StatusCode)
	}

	defer response.Body.Close()

	log.Debugf("Writing response to '%s'...\n", artifact.LocalPath)
	if _, err := io.Copy(f, response.Body); err != nil {
		return fmt.Errorf("failed to read HTTP response: %v", err)
	}

	return nil
}

func (u *SignedURL) delete(client *http.Client, artifact *Artifact) error {
	log.Debugf("DELETE '%s'...\n", u.URL)

	req, err := http.NewRequest("DELETE", u.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %v", err)
	}

	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute DELETE request: %v", err)
	}

	defer response.Body.Close()

	log.Debugf("DELETE request got %d response.\n", response.StatusCode)
	if !common.IsStatusOK(response.StatusCode) {
		return fmt.Errorf("GET failed with %d", response.StatusCode)
	}

	return nil
}

func (u *SignedURL) GetObject() (string, error) {
	URL, _ := url.Parse(u.URL)

	switch host := URL.Host; {
	case host == "storage.googleapis.com":
		log.Debugf("Parsing GCS URL: %s\n", u.URL)
		return parseGoogleStorageURL(URL)

	case strings.HasSuffix(host, "amazonaws.com"):
		log.Debug("Parsing S3 URL: %s\n", u.URL)
		return parseS3URL(URL)

	case strings.HasPrefix(host, "127.0.0.1"):
		log.Debug("Parsing localhost URL: %s\n", u.URL)
		return parseLocalhostURL(URL)

	default:
		log.Warnf("Failed to parse URL '%s' - unrecognized host '%s'\n", u.URL, host)
		return "", fmt.Errorf("unrecognized host %s", host)
	}
}

// GCS URLs follow the format 'https://storage.googleapis.com/<bucket-name>/<path>'
func parseGoogleStorageURL(URL *url.URL) (string, error) {
	re := regexp.MustCompile(`https:\/\/storage\.googleapis\.com\/[a-z0-9\-]+\/([^?]+)\?Expires=`)
	parsed := re.FindStringSubmatch(URL.String())
	if len(parsed) < 2 {
		log.Warn("Failed to parse GCS URL.\n")
		return "", fmt.Errorf("bad URL")
	}

	return parsed[1], nil
}

// S3 URLs follow the format 'https://<bucket-name>.s3.<region>.amazonaws.com/<path>'
// Note: S3 URLs use the project id as a prefix, so we take that into account here as well
func parseS3URL(URL *url.URL) (string, error) {
	re := regexp.MustCompile(`https:\/\/(.+)\.s3\.(.+)\.amazonaws\.com\/[a-z0-9\-]+\/([^?]+)\?`)
	parsed := re.FindStringSubmatch(URL.String())
	if len(parsed) < 4 {
		log.Warn("Failed to parse S3 URL.\n")
		return "", fmt.Errorf("")
	}

	return parsed[3], nil
}

// Localhost URLs are used during tests
func parseLocalhostURL(URL *url.URL) (string, error) {
	// we don't want the leading slash
	return URL.Path[1:], nil
}
