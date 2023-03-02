package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	api "github.com/semaphoreci/artifact/pkg/api"
	"github.com/semaphoreci/artifact/pkg/common"
	"github.com/semaphoreci/artifact/pkg/logger"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	URL        string
	Token      string
	HttpClient *http.Client
}

type GenerateSignedURLsRequestType int

const (
	GenerateSignedURLsRequestPUSH GenerateSignedURLsRequestType = iota
	GenerateSignedURLsRequestPUSHFORCE
	GenerateSignedURLsRequestPULL
	GenerateSignedURLsRequestYANK
)

type GenerateSignedURLsRequest struct {
	Paths []string                      `json:"paths,omitempty"`
	Type  GenerateSignedURLsRequestType `json:"type,omitempty"`
}

type GenerateSignedURLsResponse struct {
	Urls  []*api.SignedURL `json:"urls,omitempty"`
	Error string           `json:"error,omitempty"`
}

func NewClient() (*Client, error) {
	token := os.Getenv("SEMAPHORE_ARTIFACT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("SEMAPHORE_ARTIFACT_TOKEN is not set")
	}

	orgURL := os.Getenv("SEMAPHORE_ORGANIZATION_URL")
	if orgURL == "" {
		return nil, fmt.Errorf("SEMAPHORE_ORGANIZATION_URL is not set")
	}

	u, err := url.Parse(orgURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SEMAPHORE_ORGANIZATION_URL '%s': %v", orgURL, err)
	}

	u.Path = "/api/v1/artifacts"

	log.Debug("Hub client properly configured.\n")
	log.Debugf("* URL: %s\n", u.String())

	return &Client{
		URL:        u.String(),
		Token:      token,
		HttpClient: http.DefaultClient,
	}, nil
}

func (c *Client) GenerateSignedURLs(remotePaths []string, requestType GenerateSignedURLsRequestType, verbose bool) (*GenerateSignedURLsResponse, error) {
	reqBody := GenerateSignedURLsRequest{
		Paths: remotePaths,
		Type:  requestType,
	}

	log.Debug("Sending request to generate signed URLs...\n")
	log.Debugf("* Request type: %v\n", requestType)
	log.Debugf("* Paths: %v\n", remotePaths)

	var response GenerateSignedURLsResponse

	req, err := createRequest("POST", c.URL, c.Token, reqBody)
	if err != nil {
		return nil, err
	}

	retryClient := retryablehttp.NewClient()

	// 4 retries means 5 requests in total
	retryClient.RetryMax = 4
	retryClient.RetryWaitMax = 1 * time.Second
	retryClient.Logger = newLogger(verbose)

	httpResp, err := retryClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request did not return a non-5xx response: %v", err)
	}

	err = decodeResponse(httpResp, &response)
	if err != nil {
		return nil, err
	}

	log.Debug("Successfully generated signed URLs.\n")
	return &response, nil
}

func createRequest(method, url, token string, reqBody interface{}) (*retryablehttp.Request, error) {
	var serializedRequestRata bytes.Buffer
	if err := json.NewEncoder(&serializedRequestRata).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("Failed to encode http data: %v", err)
	}
	req, err := retryablehttp.NewRequest(method, url, serializedRequestRata.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Failed to create new Request: %v", err)
	}
	req.Header.Set("authorization", token)
	return req, nil
}

func decodeResponse(httpResp *http.Response, response *GenerateSignedURLsResponse) error {
	defer httpResp.Body.Close()

	if !common.IsStatusOK(httpResp.StatusCode) {
		return fmt.Errorf("failed to generate signed URLs - hub returned %d status code", httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode signed URL http response: %v", err)
	}

	if len(response.Error) > 0 {
		return fmt.Errorf("signed URL response returned errors: %s", response.Error)
	}

	return nil
}

func newLogger(verbose bool) *log.Logger {
	l := log.New()
	l.SetFormatter(new(logger.CustomFormatter))
	if verbose {
		l.SetLevel(log.DebugLevel)
	}

	return l
}
