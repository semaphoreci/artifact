package httputil

import (
	"io"
	"io/ioutil"
	"net/http"

	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	"go.uber.org/zap"
)

var httpClient = &http.Client{}

// CheckStatus checks if the status of the http response is a failure.
func CheckStatus(s int) (fail bool) {
	return s < http.StatusOK || s >= http.StatusMultipleChoices
}

// formatIfErr checks if the http result is okay, logs any errors including
// wrong status, and content in that case.
func formatIfErr(s int, descr, u string, r io.Reader) (fail bool) {
	if fail = CheckStatus(s); !fail {
		return
	}
	content, err := ioutil.ReadAll(r)
	if err != nil {
		errutil.L.Error("Failed to read http response", zap.Error(err),
			zap.String("while doing", descr), zap.String("url", u))
		return
	}
	errutil.L.Warn("http status fail", zap.String("while doing", descr), zap.String("url", u),
		zap.Int("status code", s), zap.String("status", http.StatusText(s)),
		zap.String("content", string(content)))
	return
}

// do is a shortener for http methods that can't be accessed directly in Go.
func do(descr, u, method string, content io.Reader) (fail bool) {
	req, err := http.NewRequest(http.MethodPut, u, content)
	if err != nil {
		errutil.L.Error("Failed to create new http request", zap.Error(err),
			zap.String("while doing", descr), zap.String("url", u))
		return true
	}
	res, err := httpClient.Do(req)
	defer res.Body.Close()
	return formatIfErr(res.StatusCode, "Upload", u, res.Body)
}

// UploadReader uploads content to the given signed URL.
func UploadReader(u string, content io.Reader) (fail bool) {
	return do("Upload", u, http.MethodPut, content)
}

// DownloadWriter downloads content from the given signed URL to the given io Writer.
func DownloadWriter(u string, w io.Writer) (fail bool) {
	resp, err := http.Get(u)
	if err != nil {
		errutil.L.Error("Failed to http get", zap.Error(err),
			zap.String("while doing", "Download"), zap.String("url", u))
		return true
	}
	defer resp.Body.Close()
	if fail = formatIfErr(resp.StatusCode, "Download", u, resp.Body); fail {
		return
	}
	if _, err = io.Copy(w, resp.Body); err != nil {
		errutil.L.Error("Failed to read http response", zap.Error(err),
			zap.String("while doing", "Download"), zap.String("url", u))
		return true
	}
	return
}

// DeleteURL deletes the target of the given signed URL.
func DeleteURL(u string) (fail bool) {
	return do("Delete", u, http.MethodDelete, nil)
}

// CheckURL checks if the given signed URL exists by a HEAD http request. Non-existance
// doesn't fail with an error.
func CheckURL(u string) (exist bool, fail bool) {
	resp, err := http.Head(u)
	if err != nil {
		errutil.L.Error("HEAD error", zap.String("URL", u))
		return false, true
	}
	defer resp.Body.Close()
	exist = CheckStatus(resp.StatusCode)

	return
}
