package gcs

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	httputil "github.com/semaphoreci/artifact/pkg/util/http"
	"github.com/semaphoreci/artifact/pkg/util/log"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
	"go.uber.org/zap"
)

const (
	randPostfixLen = 6
	randChars      = "abcdefghijklmnopqrstuvwxyz0123456789"
	gatewayAPIBase = "/api/v1/artifacts"
)

var (
	token      string
	client     = http.Client{}
	gatewayAPI string
)

// Init initializes Google Coud Storage with the given bucket name in environment variable.
// Loads credentials from environment variable too.
func Init() {
	rand.Seed(time.Now().UnixNano())
	token = os.Getenv("SEMAPHORE_ARTIFACT_TOKEN")
	log.Debug("initiating artifact...", zap.String("token", token))
	orgURL := os.Getenv("SEMAPHORE_ORGANIZATION_URL")
	u, err := url.Parse(orgURL)
	if err != nil {
		log.Panic("failed to parse", zap.String("org URL", orgURL), zap.Error(err))
	}
	u.Path = gatewayAPIBase
	gatewayAPI = u.String()
	log.Debug("artifact initiated", zap.String("org URL", orgURL),
		zap.String("gatewayAPIBase", gatewayAPIBase), zap.String("gatewayAPI", gatewayAPI))
}

// isFile returns if the given path points to a file in the local file system.
func isFile(filename string) (isF bool, ok bool) {
	fi, err := os.Stat(filename)
	if err == nil {
		return !fi.IsDir(), true
	}
	log.Error("Failed to find file to push", zap.String("filename", filename), zap.Error(err))
	return false, false
}

// isDir returns if the given path points to a directory in the local file system.
func isDir(filename string) (isD bool, ok bool) {
	fi, err := os.Stat(filename)
	if err == nil {
		return fi.IsDir(), true
	}
	log.Error("Failed to find dir to push", zap.String("path", filename), zap.Error(err))
	return false, false
}

// isFileSrc checks, if the given source exists, and if it's a file.
func isFileSrc(src string) (isF bool, ok bool) {
	if isF, ok = isFile(src); !ok {
		return
	}
	if isF {
		log.Debug("the source seems to be a file", zap.String("source", src))
		return
	}
	var isD bool
	if isD, ok = isDir(src); !ok {
		return
	}
	if isD {
		log.Debug("the source seems to be a directory", zap.String("source", src))
		return
	}
	errutil.ErrNotFound(src, errutil.Lfs)
	return false, false
}

// ParseURL parses object path from a signed URL.
func ParseURL(u string) string {
	if strings.HasPrefix(u, "https://storage.googleapis.com") {
		// GCS URLs follow the format 'https://storage.googleapis.com/<bucket-name>/<path>'
		re := regexp.MustCompile(`https:\/\/storage\.googleapis\.com\/[a-z0-9\-]+\/([^?]+)\?Expires=`)
		parsed := re.FindStringSubmatch(u)
		if len(parsed) < 2 {
			log.Warn("ParseURL fails to parse", zap.String("url", u))
			return ""
		}
		return parsed[1]
	} else {
		// S3 URLs follow the format 'https://<bucket-name>.s3.<region>.amazonaws.com/<path>'
		// Note: S3 URLs use the project id as a prefix, so we take that into account here as well
		re := regexp.MustCompile(`https:\/\/(.+)\.s3\.(.+)\.amazonaws\.com\/[a-z0-9\-]+\/([^?]+)\?`)
		parsed := re.FindStringSubmatch(u)
		if len(parsed) < 4 {
			log.Warn("ParseURL fails to parse", zap.String("url", u))
			return ""
		}
		return parsed[3]
	}
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
func handleHTTPReq(data interface{}, target *GenerateSignedURLsResponse) (ok bool) {
	var b bytes.Buffer
	if data != nil {
		if err := json.NewEncoder(&b).Encode(data); err != nil {
			log.VerboseError("failed to encode http data", zap.Error(err))
			return
		}
	}
	q, err := http.NewRequest(http.MethodPost, gatewayAPI, &b)
	if err != nil {
		log.VerboseError("failed to create signed URL http request", zap.Error(err))
		return
	}
	q.Header.Set("authorization", token)
	r, err := client.Do(q)
	if err != nil {
		log.VerboseError("failed to do signed URL http request", zap.Error(err))
		return
	}
	defer r.Body.Close()
	if ok = httputil.IsStatusOK(r.StatusCode); !ok {
		log.VerboseError("http do signed URL request status is an error",
			zap.Int("status code", r.StatusCode),
			zap.String("status", http.StatusText(r.StatusCode)))
		return
	}
	b.Reset()
	tee := io.TeeReader(r.Body, &b)
	if err = json.NewDecoder(tee).Decode(target); err != nil {
		log.VerboseError("failed to decode signed URL http response", zap.Error(err),
			zap.String("content", b.String()))
		return
	}
	if len(target.Error) > 0 {
		log.VerboseError("Error signed URL http result", zap.String("error", target.Error))
		return
	}
	return true
}

func randomString() string {
	output := make([]byte, randPostfixLen)
	randomness := make([]byte, randPostfixLen)

	// generate some random bytes, this shouldn't fail
	_, err := rand.Read(randomness)
	if err != nil {
		log.Error("Failed to generate random number", zap.Error(err))
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

// UploadFile uploads a file given by its filename to the Google Cloud Storage.
func UploadFile(u, filename string) (ok bool) {
	f, err := os.Open(filename)
	if err != nil {
		log.Error("Failed to open file for uploading", zap.String("filename", filename),
			zap.Error(err))
		return
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		log.Error("Failed to stat file for uploading", zap.String("filename", filename),
			zap.Error(err))
		return
	}

	return httputil.UploadReader(u, f, fileInfo.Size())
}

// PushPaths returns source and destination paths to push a file to Google Cloud Storage.
// Source path becomes a relative path on the file system, destination path becomes a category
// prefixed path to the GCS Bucket.
func PushPaths(dst, src string) (string, string) {
	newDst := pathutil.ToRelative(dst)
	newDst = pathutil.PrefixedPathFromSource(newDst, src)
	newSrc := path.Clean(src)
	log.Debug("PushPaths", zap.String("input destination", dst),
		zap.String("input source", src), zap.String("output destination", newDst),
		zap.String("output source", newSrc))
	return newDst, newSrc
}

// PushGCS uploads a file or directory from the file system to Google Cloud Storage to
// given destination. Returns if it was a success, otherwise the error has been logged.
func PushGCS(dst, src string, force bool) (ok bool) {
	log.Debug("pushing...", zap.String("source", src), zap.String("destination", dst), zap.Bool("force", force))

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
		prefLen := len(src)
		err := filepath.Walk(src, func(filename string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rps = append(rps, path.Join(dst, filename[prefLen:]))
			lps = append(lps, filename)
			return nil
		})
		if err != nil {
			log.Error("Failed to walk local directory for pushing", zap.Error(err))
			return false
		}
	}
	count := len(rps)

	request := &GenerateSignedURLsRequest{Paths: rps}
	if force {
		request.Type = generateSignedURLsRequestPUSHFORCE
	} else {
		request.Type = generateSignedURLsRequestPUSH
	}
	var t GenerateSignedURLsResponse
	ok = errutil.RetryOnFailure("get push signed URL", func() bool {
		return handleHTTPReq(request, &t)
	})
	ok = ok && doPushGCS(dst, force, count, rps, lps, t)
	if !ok {
		log.Error("File or dir not found. Please check if the source you are trying to push exists.")
		return false
	}
	return true
}

// doPushGCS does the file or directory uploading from the file system to
// Google Cloud Storage. Returns if it was a success, otherwise the error has been logged.
func doPushGCS(dst string, force bool, count int, rps, lps []string, t GenerateSignedURLsResponse) (ok bool) {
	us := t.Urls
	j := 0
	exist := false
	for i := 0; i < count; i, j = i+1, j+1 { // uploading files
		if !force { // needs to be checked if nothing exists there
			if exist, ok = httputil.CheckURL(us[j].URL); !ok {
				return
			}
			if exist {
				errutil.ErrAlreadyExists("Uploading object", lps[i], errutil.Gcs)
				return false
			}
			j++
		}

		log.Debug("Uploading...", zap.String("source", lps[i]), zap.String("destination", rps[i]))

		if ok = UploadFile(us[j].URL, lps[i]); !ok {
			return
		}
	}

	return true
}

// PullPaths returns source and destination paths to pull a file from Google Cloud Storage.
// Source path becomes a category prefixed path to the GCS Bucket,
// destination path becomes a relative path on the file system.
func PullPaths(dst, src string) (string, string) {
	newSrc := pathutil.ToRelative(src)
	newDst := pathutil.PathFromSource(dst, newSrc)
	newSrc = pathutil.PrefixedPath(newSrc)
	newDst = path.Clean(newDst)
	log.Debug("PullPaths", zap.String("input destination", dst),
		zap.String("input source", src), zap.String("output destination", newDst),
		zap.String("output source", newSrc))
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

// PullGCS downloads a file or directory from the Google Cloud Storage to the file system
// with given destination and source path.
func PullGCS(dst, src string, force bool) (ok bool) {
	log.Debug("pulling...", zap.String("source", src), zap.String("destination", dst),
		zap.Bool("force", force))
	ps := []string{src}
	var t GenerateSignedURLsResponse
	request := &GenerateSignedURLsRequest{Paths: ps, Type: generateSignedURLsRequestPULL}
	ok = errutil.RetryOnFailure("get pull signed URL", func() bool {
		return handleHTTPReq(request, &t)
	})
	ok = ok && doPullGCS(dst, src, force, t)
	if !ok {
		log.Error("Artifact not found. Please check if the artifact you are trying to pull exists.")
		return false
	}
	return true
}

// doPullGCS does the file downloading from the given signed URLs.
func doPullGCS(dst, src string, force bool, t GenerateSignedURLsResponse) (ok bool) {
	if len(t.Urls) == 1 { // one file only
		url := t.Urls[0].URL
		obj := ParseURL(url)
		// removing <project-name>/<category>/<projectID>/ prefix
		obj = cutPrefixByDelimMulti(obj, '/', 3)
		if obj == src { // they are the same: requested single file pull
			return PullFileGCS(dst, url, force)
		} // otherwise it will be downloaded as a directory
	}
	prefLen := len(src)
	for _, u := range t.Urls { // iterate all urls and put them in a directory structure
		obj := ParseURL(u.URL)
		if ok = PullFileGCS(path.Join(dst, obj[prefLen:]), u.URL, force); !ok {
			return
		}
	}
	return true
}

// PullFileGCS downloads a file from the Google Cloud Storage to the file system with given
// source path.
func PullFileGCS(dstFilename, u string, force bool) (ok bool) {
	log.Debug("downloading...", zap.String("url", u), zap.String("destination", dstFilename))
	if !force {
		if _, err := os.Stat(dstFilename); err == nil {
			errutil.ErrAlreadyExists("Downloading file", dstFilename, errutil.Lfs)
			return
		}
	}
	err := os.MkdirAll(filepath.Dir(dstFilename), 0755)
	if err != nil {
		log.Error("Failed to create dir for pulling from Google Cloud Storage", zap.Error(err))
		return
	}
	var f *os.File
	if f, err = os.Create(dstFilename); err != nil {
		log.Error("Failed to create file for pulling from Google Cloud Storage", zap.Error(err))
		return
	}
	defer f.Close()
	ok = httputil.DownloadWriter(u, f)
	log.Debug("PullFileGCS result", zap.Bool("success", ok))
	return ok
}

// YankPath returns path to yank a file from Google Cloud Storage.
// Path becomes a category prefixed path to the GCS Bucket.
func YankPath(f string) string {
	newF := pathutil.ToRelative(f)
	newF = pathutil.PrefixedPath(newF)
	log.Debug("YankPath", zap.String("input file", f), zap.String("output file", newF))
	return newF
}

// YankGCS deletes a file or directory from the Google Cloud Storage.
func YankGCS(name string) (ok bool) {
	ps := []string{name}
	var t GenerateSignedURLsResponse
	request := &GenerateSignedURLsRequest{Paths: ps, Type: generateSignedURLsRequestYANK}
	ok = errutil.RetryOnFailure("get yank signed URL", func() bool {
		return handleHTTPReq(request, &t)
	})
	ok = ok && doYankGCS(t)
	if !ok {
		log.Warn("Artifact not found. Please check if the artifact you are trying to yank exists.")
		return false
	}
	return true
}

func doYankGCS(t GenerateSignedURLsResponse) (ok bool) {
	if len(t.Urls) == 1 {
		if ok = httputil.DeleteURL(t.Urls[0].URL); !ok {
			return false
		}
	} else {
		for _, u := range t.Urls {
			if ok = httputil.DeleteURL(u.URL); !ok {
				return false
			}
		}
	}
	return true
}
