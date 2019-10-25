package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	httputil "github.com/semaphoreci/artifact/pkg/util/http"
	humeutil "github.com/semaphoreci/artifact/pkg/util/hume"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
)

const (
	randPostfixLen = 6
	randChars      = "abcdefghijklmnopqrstuvwxyz0123456789"
	gatewayAPIBase = "/api/v1/artifacts"
)

var (
	ctx        = context.Background()
	token      string
	client     = http.Client{}
	re         = regexp.MustCompile(`https:\/\/storage\.googleapis\.com\/[a-z0-9\-]+\/([^?]+)\?Expires=`)
	gatewayAPI string
)

// init initializes Google Coud Storage with the given bucket name in environment variable.
// Loads credentials from environment variable too.
func init() {
	token = os.Getenv("SEMAPHORE_ARTIFACT_TOKEN")
	errutil.Debug("artifact inited with token name: %s", token)
	orgURL := os.Getenv("SEMAPHORE_ORGANIZATION_URL")
	u, err := url.Parse(orgURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse organization url: '%s', err: %v", u, err))
	}
	u.Path = gatewayAPIBase
	gatewayAPI = u.String()
	errutil.Debug("orgURL: %s, gatewayAPIBase: %s, gatewayAPI: %s", orgURL, gatewayAPIBase, gatewayAPI)
}

// isFile returns if the given path points to a file in the local file system.
func isFile(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		return !fi.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// isDir returns if the given path points to a directory in the local file system.
func isDir(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		return fi.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// isFileSrc checks, if the given source exists, and if it's a file.
func isFileSrc(src string) (bool, error) {
	ok, err := isFile(src)
	if err != nil {
		errutil.Info("isFileSrc '%s' gives an error: %s", src, err)
		return false, err
	}
	if ok {
		errutil.Info("isFileSrc '%s' says: it's a file", src)
		return true, nil
	}
	if ok, err = isDir(src); err != nil {
		errutil.Info("isFileSrc '%s' gives an error: %s", src, err)
		return false, err
	}
	if ok {
		errutil.Info("isFileSrc '%s' says: it's a directory", src)
		return false, nil
	}
	return false, &errutil.ErrNotFound{src, errutil.Lfs}
}

// ParseURL parses object path from a signed URL.
func ParseURL(u string) string {
	parsed := re.FindStringSubmatch(u)
	if len(parsed) < 2 {
		errutil.Warn("parsing URL", fmt.Errorf("ParseURL fails to parse '%s'", u))
		return ""
	}
	return parsed[1]
}

// SignedURL contains an url and its method type.
type SignedURL struct {
	URL    string `json:"url,omitempty"`
	Method string `json:"method,omitempty"`
}

type generateSignedURLsRequestType int

const (
	generateSignedURLsRequestPUSH generateSignedURLsRequestType = iota
	generateSignedURLsRequestPUSHFORCE
	generateSignedURLsRequestPULL
	generateSignedURLsRequestYANK
)

// GenerateSignedURLsRequest is the request for Push call. Contains a list of paths to upload,
// and if it's forced.
type GenerateSignedURLsRequest struct {
	Paths []string                      `json:"paths,omitempty"`
	Type  generateSignedURLsRequestType `json:"type,omitempty"`
}

// GenerateSignedURLsResponse contain a list of Signed URLs. It can be used for multiple grcp calls.
type GenerateSignedURLsResponse struct {
	Urls  []*SignedURL `json:"urls,omitempty"`
	Error string       `json:"error,omitempty"`
}

func handleHTTPReq(data interface{}, target *GenerateSignedURLsResponse) error {
	var b bytes.Buffer
	if data != nil {
		err := json.NewEncoder(&b).Encode(data)
		if err != nil {
			return errutil.Error("failed to encode data", err)
		}
	}
	q, err := http.NewRequest(http.MethodPost, gatewayAPI, &b)
	if err != nil {
		return errutil.Error("failed to create request", err)
	}
	q.Header.Set("authorization", token)
	r, err := client.Do(q)
	if err != nil {
		return errutil.Error("failed to do request", err)
	}
	defer r.Body.Close()
	b.Reset()
	tee := io.TeeReader(r.Body, &b)
	if err = json.NewDecoder(tee).Decode(target); err != nil {
		err := fmt.Errorf("failed to decode response: %s, content: %s", err, b.String())
		return errutil.Error("decoding response", err)
	}
	if len(target.Error) > 0 {
		return errutil.Error("Error http result", errors.New(target.Error))
	}
	return nil
}

func retryableHTTPReq(data interface{}, target *GenerateSignedURLsResponse) error {
	retries := 0
	maxRetries := 3

	for {
		err := handleHTTPReq(data, target)
		if err == nil {
			return nil
		}
		retries++
		if retries <= maxRetries {
			errutil.Info("Failed to perform request, retrying in 3 sec (%d out of %d).", retries, maxRetries)
			time.Sleep(3 * time.Second)
			continue
		} else {
			errutil.Info("Give up after %d retries.", maxRetries)
			return err
		}
	}
}

func randomString() (string, error) {
	output := make([]byte, randPostfixLen)
	randomness := make([]byte, randPostfixLen)

	// generate some random bytes, this shouldn't fail
	_, err := rand.Read(randomness)
	if err != nil {
		err = errutil.Error("Generating random number", err)
		return "", err
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

// CreateExpireFileName creates a new name for an expire descriptor file on the Google Cloud Storage.
func CreateExpireFileName(expTime time.Duration) string {
	if expTime < 1 {
		return ""
	}

	randPostfix, err := randomString()
	if err != nil {
		return ""
	}
	expFilename := strconv.FormatInt(time.Now().Add(expTime).Unix(), 10)
	var b strings.Builder
	expFilename = path.Join(pathutil.ExpirePrefix, expFilename)
	b.WriteString(expFilename)
	b.WriteByte('-')
	b.WriteString(randPostfix)
	expFilename = b.String()
	errutil.Debug("CreateExpireFileName succeeded for exptime: '%s', result: '%s", expTime, expFilename)
	return expFilename
}

// UploadFile uploads a file given by its filename to the Google Cloud Storage.
func UploadFile(u, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	err = httputil.UploadReader(u, f)
	errutil.Debug("Uploading file '%s', (optional) error: %v", filename, err)
	return err
}

// PushPaths returns source and destination paths to push a file to Google Cloud Storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path to the GCS Bucket.
func PushPaths(dst, src string) (string, string) {
	newDst := pathutil.ToRelative(dst)
	newDst = pathutil.PrefixedPathFromSource(newDst, src)
	newSrc := path.Clean(src)
	errutil.Debug("PushPaths input dst: '%s', src: '%s', output dst: '%s', src: '%s'", dst, src,
		newDst, newSrc)
	return newDst, newSrc
}

// PushGCS uploads a file or directory from the file system to Google Cloud Storage to given destination
// with a human readable expire string. Returns expire filename and error, if happened any.
func PushGCS(dst, src, expires string, force bool) error {
	errutil.Debug("pushing '%s' to '%s' with force=%t", src, dst, force)
	expTime, err := humeutil.ParseRelativeAgeForHumans(expires)
	if err != nil {
		return err
	}

	isFile, err := isFileSrc(src)
	if err != nil {
		return err
	}

	// local and remote paths
	var lps, rps []string

	if isFile {
		rps = []string{dst}
		lps = []string{src}
	} else { // directory, getting all filenames
		rps = []string{}
		lps = []string{}
		prefLen := len(src)
		if err = filepath.Walk(src, func(filename string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rps = append(rps, path.Join(dst, filename[prefLen:]))
			lps = append(lps, filename)
			return nil
		}); err != nil {
			return errutil.Error("PushGCS walk failed", err)
		}
	}
	count := len(rps) // uploading files until this, expire file works differently
	if expTime > 0 {
		rps = append(rps, CreateExpireFileName(expTime))
	}

	request := &GenerateSignedURLsRequest{Paths: rps}
	if force {
		request.Type = generateSignedURLsRequestPUSHFORCE
	} else {
		request.Type = generateSignedURLsRequestPUSH
	}
	var t GenerateSignedURLsResponse
	if err = retryableHTTPReq(request, &t); err != nil {
		return err
	}
	us := t.Urls
	j := 0
	ok := false
	for i := 0; i < count; i, j = i+1, j+1 { // uploading files
		if !force { // needs to be checked if nothing exists there
			if ok, err = httputil.CheckURL(us[j].URL); err != nil {
				return err
			}
			if ok {
				err = errutil.Error("Uploading object to Google Cloud Storage",
					&errutil.ErrAlreadyExists{lps[i], errutil.Gcs})
				return err
			}
			j++
		}
		errutil.Debug("uploading '%s' to '%s'", lps[i], rps[i])
		if err = UploadFile(us[j].URL, lps[i]); err != nil {
			return err
		}
	}

	if expTime > 0 {
		if !force {
			count = count*2 + 1
		}
		if err = httputil.UploadReader(us[count].URL, strings.NewReader(dst)); err != nil {
			return err
		}
	}

	return nil
}

// PullPaths returns source and destination paths to pull a file from Google Cloud Storage.
// Source path becomes a category prefixed path to the GCS Bucket,
// destination path becomes a relative path on the file system.
func PullPaths(dst, src string) (string, string) {
	newSrc := pathutil.ToRelative(src)
	newDst := pathutil.PathFromSource(dst, newSrc)
	newSrc = pathutil.PrefixedPath(newSrc)
	newDst = path.Clean(newDst)
	errutil.Debug("PullPaths input dst: '%s', src: '%s', output dst: '%s', src: '%s'", dst, src,
		newDst, newSrc)
	return newDst, newSrc
}

// PullGCS downloads a file or directory from the Google Cloud Storage to the file system
// with given destination and source path.
func PullGCS(dst, src string, force bool) error {
	errutil.Debug("pulling '%s' to '%s' with force=%t", src, dst, force)
	ps := []string{src}
	var t GenerateSignedURLsResponse
	request := &GenerateSignedURLsRequest{Paths: ps, Type: generateSignedURLsRequestPULL}
	err := retryableHTTPReq(request, &t)
	if err != nil {
		return err
	}
	if len(t.Urls) == 1 {
		if err = PullFileGCS(dst, t.Urls[0].URL, force); err != nil {
			return err
		}
	} else {
		prefLen := len(src)
		for _, u := range t.Urls {
			obj := ParseURL(u.URL)
			if err = PullFileGCS(path.Join(dst, obj[prefLen:]), u.URL, force); err != nil {
				return err
			}
		}
	}
	return nil
}

// PullFileGCS downloads a file from the Google Cloud Storage to the file system with given source path.
func PullFileGCS(dstFilename, u string, force bool) error {
	errutil.Debug("downloading from url '%s' to '%s'", u, dstFilename)
	if !force {
		if _, err := os.Stat(dstFilename); err == nil {
			return errutil.Error("Downloading file from Google Cloud Storage",
				&errutil.ErrAlreadyExists{dstFilename, errutil.Lfs})
		}
	}
	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		return errutil.Error("Creating directory for pulling from Google Cloud Storage", err)
	}
	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		return errutil.Error("Creating result file for pulling from Google Cloud Storage", err)
	}
	defer f.Close()
	err = httputil.DownloadWriter(u, f)
	errutil.Debug("PullFileGCS result: %v", err)
	return err
}

// YankPath returns path to yank a file from Google Cloud Storage.
// Path becomes a category prefixed path to the GCS Bucket.
func YankPath(f string) string {
	newF := pathutil.ToRelative(f)
	errutil.Debug("YankPath input f: '%s', output f: '%s'", f, newF)
	return pathutil.PrefixedPath(newF)
}

// YankGCS deletes a file or directory from the Google Cloud Storage.
func YankGCS(name string) error {
	ps := []string{name}
	var t GenerateSignedURLsResponse
	request := &GenerateSignedURLsRequest{Paths: ps, Type: generateSignedURLsRequestYANK}
	err := retryableHTTPReq(request, &t)
	if err != nil {
		return err
	}
	if len(t.Urls) == 1 {
		if err = httputil.DeleteURL(t.Urls[0].URL); err != nil {
			return err
		}
	} else {
		for _, u := range t.Urls {
			if err = httputil.DeleteURL(u.URL); err != nil {
				return err
			}
		}
	}
	return nil
}
