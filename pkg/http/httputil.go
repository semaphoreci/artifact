package httputil

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var httpClient = &http.Client{}

func IsStatusOK(s int) bool {
	return s >= http.StatusOK && s < http.StatusMultipleChoices
}

// Logs more information (status, response body) in case of a failed HTTP request.
func formatIfErr(s int, descr, u string, r io.Reader) (ok bool) {
	if ok = IsStatusOK(s); ok {
		return
	}

	content, err := ioutil.ReadAll(r)
	if err != nil {
		log.Errorf("Failed to read http response: %v", err)
		return
	}

	log.Warn("HTTP request to '%s' failed with '%s': %s\n", u, s, string(content))
	return
}

// do does httpclient.Do with the given paramters. If getBody is true, the response body
// is returned, and the responsability of closing it is transferred.
func do(descr, u, method string, content io.Reader, size int64, getBody bool) (ok bool, body io.ReadCloser) {
	contentBody := content

	// If the file has no bytes, we need to use http.NoBody
	// See https://cs.opensource.google/go/go/+/refs/tags/go1.18.2:src/net/http/request.go;l=920
	if method == http.MethodPut && size == 0 {
		contentBody = http.NoBody
	}

	req, err := http.NewRequest(method, u, contentBody)
	if err != nil {
		log.Errorf("Failed to create new http request: %v\n", err)
		return
	}

	req.ContentLength = size

	res, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("Failed to execute http request: %v", err)
		return
	}

	if !getBody {
		defer res.Body.Close()
		return formatIfErr(res.StatusCode, method, u, res.Body), nil
	}
	return formatIfErr(res.StatusCode, method, u, res.Body), res.Body
}

func UploadReader(u string, content io.Reader, size int64) error {
	ok, _ := do("Upload", u, http.MethodPut, content, size, false)
	if !ok {
		return fmt.Errorf("error uploading")
	}

	return nil
}

func DownloadWriter(u string, w io.Writer) error {
	ok, body := do("Download", u, http.MethodGet, nil, 0, true)

	defer body.Close()

	if !ok {
		return fmt.Errorf("error downloading")
	}

	if _, err := io.Copy(w, body); err != nil {
		return fmt.Errorf("failed to read http response: %v", err)
	}

	return nil
}

// DeleteURL deletes the target of the given signed URL.
func DeleteURL(u string) (ok bool) {
	ok, _ = do("Delete", u, http.MethodDelete, nil, 0, false)
	return
}

// Checks if the given signed URL exists by executing a HEAD http request.
// A file not existing doesn't fail with an error.
func Exists(u string) (bool, error) {
	log.Debugf("HEAD '%s'...\n", u)

	resp, err := http.Head(u)
	if err != nil {
		return false, fmt.Errorf("error executing HEAD '%s': %v", u, err)
	}

	defer resp.Body.Close()

	return IsStatusOK(resp.StatusCode), nil
}
