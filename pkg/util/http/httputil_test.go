package httputil

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	httpmock "github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

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

func TestHTTPMethod(t *testing.T) {
	descr := "some description"
	u := "example.com"
	content := "some content"
	cr := strings.NewReader(content)

	check := func(method string, cr io.Reader, getBody bool) {
		httpmock.ActivateNonDefault(httpClient)
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(method, u,
			func(req *http.Request) (*http.Response, error) {
				if req.Method != method {
					return httpmock.NewJsonResponse(403, "method doesn't match")
				}
				if getBody {
					bodyContent, _ := ioutil.ReadAll(req.Body)
					return httpmock.NewBytesResponse(200, bodyContent), nil
				}
				return httpmock.NewStringResponse(200, ""), nil
			},
		)
		ok, body := do(descr, u, method, cr, getBody)
		assert.True(t, ok, "should be fine but fails")
		if getBody {
			defer body.Close()
			bodyContent, err := ioutil.ReadAll(body)
			assert.Nil(t, err, "failed to read body")
			assert.Equal(t, content, string(bodyContent), "contents should match")
		}
	}

	check(http.MethodGet, cr, true)
	check(http.MethodPut, cr, false)
	check(http.MethodDelete, nil, false)
}
