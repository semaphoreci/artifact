package storage

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/semaphoreci/artifact/pkg/common"
	log "github.com/sirupsen/logrus"
)

func newHTTPClient() *retryablehttp.Client {
	return &retryablehttp.Client{
		HTTPClient:   http.DefaultClient,
		RetryWaitMin: 500 * time.Millisecond,
		RetryWaitMax: time.Second,
		RetryMax:     4,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
		ResponseLogHook: func(l retryablehttp.Logger, r *http.Response) {
			if common.IsStatusOK(r.StatusCode) {
				return
			}

			// 404 on a HEAD request is not an error for us.
			if r.StatusCode == http.StatusNotFound && r.Request.Method == http.MethodHead {
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Errorf(
					"%s request to %s failed with %d status code\n",
					r.Request.Method,
					r.Request.URL,
					r.StatusCode,
				)
			}

			log.Errorf(
				"%s request to %s failed with %d status code: %s\n",
				r.Request.Method,
				r.Request.URL,
				r.StatusCode,
				string(body),
			)
		},
	}
}
