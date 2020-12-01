package gcs

import (
	"bytes"
	"context"
	"encoding/json"
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
	"go.uber.org/zap"
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
	gatewayAPI string
)

// init initializes Google Coud Storage with the given bucket name in environment variable.
// Loads credentials from environment variable too.
func init() {
	token = os.Getenv("SEMAPHORE_ARTIFACT_TOKEN")
	errutil.L.Debug("initiating artifact...", zap.String("token", token))
	orgURL := os.Getenv("SEMAPHORE_ORGANIZATION_URL")
	u, err := url.Parse(orgURL)
	if err != nil {
		errutil.L.Panic("failed to parse", zap.String("org URL", orgURL), zap.Error(err))
	}
	u.Path = gatewayAPIBase
	gatewayAPI = u.String()
	errutil.L.Debug("artifact initiated", zap.String("org URL", orgURL),
		zap.String("gatewayAPIBase", gatewayAPIBase), zap.String("gatewayAPI", gatewayAPI))
}

// isFile returns if the given path points to a file in the local file system.
func isFile(filename string) (isF bool, fail bool) {
	fi, err := os.Stat(filename)
	if err == nil {
		return !fi.IsDir(), false
	}
	if os.IsNotExist(err) {
		return false, false
	}
	errutil.L.Error("looking for a file failed", zap.Error(err))
	return false, true
}

// isDir returns if the given path points to a directory in the local file system.
func isDir(filename string) (isD bool, fail bool) {
	fi, err := os.Stat(filename)
	if err == nil {
		return fi.IsDir(), false
	}
	if os.IsNotExist(err) {
		return false, false
	}
	errutil.L.Error("looking for a directory failed", zap.Error(err))
	return false, true
}

// isFileSrc checks, if the given source exists, and if it's a file.
func isFileSrc(src string) (isF bool, fail bool) {
	if isF, fail = isFile(src); fail {
		return
	}
	if isF {
		errutil.L.Debug("the source seems to be a file", zap.String("source", src))
		return
	}
	var isD bool
	if isD, fail = isDir(src); fail {
		return false, true
	}
	if isD {
		errutil.L.Debug("the source seems to be a directory", zap.String("source", src))
		return false, false
	}
	errutil.L.Error("the source seems to be a directory", zap.String("source", src))
	errutil.ErrNotFound(src, errutil.Lfs)
	return false, true
}

// ParseURL parses object path from a signed URL.
func ParseURL(u string) string {
	re := regexp.MustCompile(`https:\/\/storage\.googleapis\.com\/[a-z0-9\-]+\/([^?]+)\?Expires=`)
	parsed := re.FindStringSubmatch(u)
	if len(parsed) < 2 {
		errutil.L.Warn("ParseURL fails to parse", zap.String("url", u))
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

// GenerateSignedURLsResponse contain a list of Signed URLs. It can be used for
// multiple grcp calls.
type GenerateSignedURLsResponse struct {
	Urls  []*SignedURL `json:"urls,omitempty"`
	Error string       `json:"error,omitempty"`
}

// handleHTTPReq calls a signed url, and returns response in arg pointer.
func handleHTTPReq(data interface{}, target *GenerateSignedURLsResponse) (fail bool) {
	var b bytes.Buffer
	fail = true
	if data != nil {
		if err := json.NewEncoder(&b).Encode(data); err != nil {
			errutil.L.Error("failed to encode http data", zap.Error(err))
			return
		}
	}
	q, err := http.NewRequest(http.MethodPost, gatewayAPI, &b)
	if err != nil {
		errutil.L.Error("failed to create signed URL http request", zap.Error(err))
		return
	}
	q.Header.Set("authorization", token)
	r, err := client.Do(q)
	if err != nil {
		errutil.L.Error("failed to do signed URL http request", zap.Error(err))
		return
	}
	defer r.Body.Close()
	if fail = httputil.CheckStatus(r.StatusCode); fail {
		errutil.L.Error("http do signed URL request status is an error",
			zap.Int("status code", r.StatusCode),
			zap.String("status", http.StatusText(r.StatusCode)))
		return
	}
	b.Reset()
	tee := io.TeeReader(r.Body, &b)
	if err = json.NewDecoder(tee).Decode(target); err != nil {
		errutil.L.Error("failed to decode signed URL http response", zap.Error(err),
			zap.String("content", b.String()))
		return
	}
	if len(target.Error) > 0 {
		errutil.L.Error("Error signed URL http result", zap.String("error", target.Error))
		return
	}
	return false
}

func randomString() string {
	output := make([]byte, randPostfixLen)
	randomness := make([]byte, randPostfixLen)

	// generate some random bytes, this shouldn't fail
	_, err := rand.Read(randomness)
	if err != nil {
		errutil.L.Error("Failed to generate random number", zap.Error(err))
		return ""
	}

	// fill output
	l := uint8(len(randChars))
	for pos := 0; pos < randPostfixLen; pos++ {
		random := uint8(randomness[pos])   // get random item
		randomPos := random % uint8(l)     // random % length
		output[pos] = randChars[randomPos] // put into output
	}
	return string(output)
}

// CreateExpireFileName creates a new name for an expire descriptor file on the
// Google Cloud Storage.
func CreateExpireFileName(expTime time.Duration) string {
	if expTime < 1 {
		return ""
	}

	randPostfix := randomString()
	if len(randPostfix) == 0 {
		return ""
	}
	expFilename := strconv.FormatInt(time.Now().Add(expTime).Unix(), 10)
	var b strings.Builder
	expFilename = path.Join(pathutil.ExpirePrefix, expFilename)
	b.WriteString(expFilename)
	b.WriteByte('-')
	b.WriteString(randPostfix)
	expFilename = b.String()
	errutil.L.Debug("CreateExpireFileName succeeded", zap.Duration("expire time", expTime),
		zap.String("result", expFilename))
	return expFilename
}

// UploadFile uploads a file given by its filename to the Google Cloud Storage.
func UploadFile(u, filename string) (fail bool) {
	f, err := os.Open(filename)
	if err != nil {
		errutil.L.Error("Failed to open file for uploading", zap.String("filename", filename),
			zap.Error(err))
		return true
	}
	defer f.Close()
	return httputil.UploadReader(u, f)
}

// PushPaths returns source and destination paths to push a file to Google Cloud Storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path to the GCS Bucket.
func PushPaths(dst, src string) (string, string) {
	newDst := pathutil.ToRelative(dst)
	newDst = pathutil.PrefixedPathFromSource(newDst, src)
	newSrc := path.Clean(src)
	errutil.L.Debug("PushPaths", zap.String("input destination", dst),
		zap.String("input source", src), zap.String("output destination", newDst),
		zap.String("output source", newSrc))
	return newDst, newSrc
}

// PushGCS uploads a file or directory from the file system to Google Cloud Storage to
// given destination with a human readable expire string. Returns if happened any error,
// that was logged in that case.
func PushGCS(dst, src, expires string, force bool) (fail bool) {
	errutil.L.Debug("pushing...", zap.String("source", src), zap.String("destination", dst),
		zap.Bool("force", force))
	expTime := humeutil.ParseRelativeAgeForHumans(expires)
	if expTime == 0 {
		return true
	}

	var isF bool
	if isF, fail = isFileSrc(src); fail {
		return
	}

	// local and remote paths
	var lps, rps []string

	if isF {
		rps = []string{dst}
		lps = []string{src}
	} else { // directory, getting all filenames
		rps = []string{}
		lps = []string{}
		prefLen := len(src)
		if err := filepath.Walk(src, func(filename string, info os.FileInfo, err error) error {
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
			errutil.L.Error("failed to walk local directory for pushing", zap.Error(err))
			return true
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
	if fail = errutil.RetryOnFailure("get push signed URL", func() bool {
		return handleHTTPReq(request, &t)
	}); fail {
		return
	}
	us := t.Urls
	j := 0
	exist := false
	for i := 0; i < count; i, j = i+1, j+1 { // uploading files
		if !force { // needs to be checked if nothing exists there
			if exist, fail = httputil.CheckURL(us[j].URL); fail {
				return
			}
			if exist {
				errutil.ErrAlreadyExists("Uploading object", lps[i], errutil.Gcs)
				return true
			}
			j++
		}
		errutil.L.Debug("Uploading...", zap.String("source", lps[i]),
			zap.String("destination", rps[i]))
		if fail = UploadFile(us[j].URL, lps[i]); fail {
			return
		}
	}

	if expTime > 0 {
		if !force {
			count = count*2 + 1
		}
		if fail = httputil.UploadReader(us[count].URL, strings.NewReader(dst)); fail {
			return
		}
	}

	return
}

// PullPaths returns source and destination paths to pull a file from Google Cloud Storage.
// Source path becomes a category prefixed path to the GCS Bucket,
// destination path becomes a relative path on the file system.
func PullPaths(dst, src string) (string, string) {
	newSrc := pathutil.ToRelative(src)
	newDst := pathutil.PathFromSource(dst, newSrc)
	newSrc = pathutil.PrefixedPath(newSrc)
	newDst = path.Clean(newDst)
	errutil.L.Debug("PullPaths", zap.String("input destination", dst),
		zap.String("input source", src), zap.String("output destination", newDst),
		zap.String("output source", newSrc))
	return newDst, newSrc
}

// PullGCS downloads a file or directory from the Google Cloud Storage to the file system
// with given destination and source path.
func PullGCS(dst, src string, force bool) (fail bool) {
	errutil.L.Debug("pulling...", zap.String("source", src), zap.String("destination", dst),
		zap.Bool("force", force))
	ps := []string{src}
	var t GenerateSignedURLsResponse
	request := &GenerateSignedURLsRequest{Paths: ps, Type: generateSignedURLsRequestPULL}
	if fail = errutil.RetryOnFailure("get pull signed URL", func() bool {
		return handleHTTPReq(request, &t)
	}); fail {
		return
	}
	if len(t.Urls) == 1 {
		if fail = PullFileGCS(dst, t.Urls[0].URL, force); fail {
			return
		}
	} else {
		prefLen := len(src)
		for _, u := range t.Urls {
			obj := ParseURL(u.URL)
			if fail = PullFileGCS(path.Join(dst, obj[prefLen:]), u.URL, force); fail {
				return
			}
		}
	}
	return
}

// PullFileGCS downloads a file from the Google Cloud Storage to the file system with given
// source path.
func PullFileGCS(dstFilename, u string, force bool) (fail bool) {
	errutil.L.Debug("downloading...", zap.String("url", u), zap.String("destination", dstFilename))
	if !force {
		if _, err := os.Stat(dstFilename); err == nil {
			errutil.ErrAlreadyExists("Downloading file", dstFilename, errutil.Lfs)
			return true
		}
	}
	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		errutil.L.Error("Creating directory for pulling from Google Cloud Storage", zap.Error(err))
		return true
	}
	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		errutil.L.Error("Creating file for pulling from Google Cloud Storage", zap.Error(err))
		return true
	}
	defer f.Close()
	fail = httputil.DownloadWriter(u, f)
	errutil.L.Debug("PullFileGCS result", zap.Bool("success", !fail))
	return
}

// YankPath returns path to yank a file from Google Cloud Storage.
// Path becomes a category prefixed path to the GCS Bucket.
func YankPath(f string) string {
	newF := pathutil.ToRelative(f)
	errutil.L.Debug("YankPath", zap.String("input file", f), zap.String("output file", newF))
	return pathutil.PrefixedPath(newF)
}

// YankGCS deletes a file or directory from the Google Cloud Storage.
func YankGCS(name string) (fail bool) {
	ps := []string{name}
	var t GenerateSignedURLsResponse
	request := &GenerateSignedURLsRequest{Paths: ps, Type: generateSignedURLsRequestYANK}
	if fail = errutil.RetryOnFailure("get yank signed URL", func() bool {
		return handleHTTPReq(request, &t)
	}); fail {
		return
	}
	if len(t.Urls) == 1 {
		if fail = httputil.DeleteURL(t.Urls[0].URL); fail {
			return
		}
	} else {
		for _, u := range t.Urls {
			if fail = httputil.DeleteURL(u.URL); fail {
				return
			}
		}
	}
	return
}
