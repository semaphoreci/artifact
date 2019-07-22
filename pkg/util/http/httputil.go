package httputil

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	errutil "github.com/semaphoreci/artifact/pkg/util/err"
)

var httpClient = &http.Client{}

// checkStatusReadAll checks if the status of the http response is OK. If it's not, tries to download
// content, and returns it.
func checkStatusReadAll(s int, r io.Reader) (bool, string, error) {
	statusOk := s >= http.StatusOK && s < http.StatusMultipleChoices
	if statusOk {
		return true, "", nil
	}
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return false, "", errutil.Error("Failed to read http response", err)
	}
	return false, string(content), nil
}

// formatIfErr checks if the http result is okay, logs and returns any errors including wrong status,
// and content in that case.
func formatIfErr(s int, descr, u string, r io.Reader) error {
	statusOk, content, err := checkStatusReadAll(s, r)
	if statusOk {
		return nil
	}
	if err != nil {
		return err
	}
	err = fmt.Errorf("%s status: %d, content: %s, url: %s", descr, s, content, u)
	return errutil.Warn("http fail", err)
}

// do is a shortener for http methods that can't be accessed directly in Go.
func do(req *http.Request, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	return httpClient.Do(req)
}

// UploadReader uploads content to the given signed URL.
func UploadReader(u string, content io.Reader) error {
	resp, err := do(http.NewRequest(http.MethodPut, u, content))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return formatIfErr(resp.StatusCode, "Upload", u, resp.Body)
}

// DownloadWriter downloads content from the given signed URL to the given io Writer.
func DownloadWriter(u string, w io.Writer) error {
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err = formatIfErr(resp.StatusCode, "Download", u, resp.Body); err != nil {
		return err
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

// DeleteURL deletes the target of the given signed URL.
func DeleteURL(u string) error {
	resp, err := do(http.NewRequest(http.MethodDelete, u, nil))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return formatIfErr(resp.StatusCode, "Delete", u, resp.Body)
}

// CheckURL checks if the given signed URL exists by a HEAD http request. Non-existance doesn't
// fail with an error.
func CheckURL(u string) (bool, error) {
	resp, err := http.Head(u)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	statusOk, content, err := checkStatusReadAll(resp.StatusCode, resp.Body)
	if err != nil {
		return false, err
	}
	errutil.Debug("HEAD result content: '%s' for url: '%s', status: %d", content, u, resp.StatusCode)

	return statusOk, nil
}
