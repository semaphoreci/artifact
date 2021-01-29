package httputil

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	httpClient = &MockClient{}
}

// MockClient is mocking a http client.
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the mock client's `Do` func.
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestDo(t *testing.T) {
	var lastReq *http.Request

	httpClient.(*MockClient).DoFunc = func(req *http.Request) (*http.Response, error) {
		lastReq = req
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(""))),
			StatusCode: http.StatusOK}, nil
	}

	check := func(result bool, expMethod string) {
		assert.Equal(t, expMethod, lastReq.Method)
		assert.Equal(t, true, result)
	}

	u := "https://example.com/some_path"
	content := "some_content"
	contentReader := strings.NewReader(content)
	res := UploadReader(u, contentReader)
	check(res, http.MethodPut)
	contentWriter := bytes.NewBufferString(content)
	res = DownloadWriter(u, contentWriter)
	check(res, http.MethodGet)
	res = DeleteURL(u)
	check(res, http.MethodDelete)
}

func TestIsStatusOK(t *testing.T) {
	check := func(statusCode int, exp bool) {
		ok := IsStatusOK(statusCode)
		assert.Equal(t, exp, ok, statusCode)
	}

	check(http.StatusContinue, false)
	check(http.StatusProcessing, false)
	check(http.StatusOK, true)
	check(http.StatusAccepted, true)
	check(http.StatusMultiStatus, true)
	check(http.StatusProcessing, false)
	check(http.StatusMovedPermanently, false)
	check(http.StatusSeeOther, false)
	check(http.StatusNotModified, false)
	check(http.StatusBadRequest, false)
	check(http.StatusUnauthorized, false)
	check(http.StatusInternalServerError, false)
	check(http.StatusNotImplemented, false)
	check(http.StatusBadGateway, false)
}
