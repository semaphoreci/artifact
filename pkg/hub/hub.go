package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	api "github.com/semaphoreci/artifact/pkg/api"
	common "github.com/semaphoreci/artifact/pkg/common"
	retry "github.com/semaphoreci/artifact/pkg/retry"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	URL        string
	Token      string
	HttpClient *http.Client
}

type generateSignedURLsRequestType int

const (
	GenerateSignedURLsRequestPUSH generateSignedURLsRequestType = iota
	GenerateSignedURLsRequestPUSHFORCE
	GenerateSignedURLsRequestPULL
	GenerateSignedURLsRequestYANK
)

type GenerateSignedURLsRequest struct {
	Paths []string                      `json:"paths,omitempty"`
	Type  generateSignedURLsRequestType `json:"type,omitempty"`
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
	log.Debugf("> URL: %s\n", u.String())

	return &Client{
		URL:        u.String(),
		Token:      token,
		HttpClient: http.DefaultClient,
	}, nil
}

func (c *Client) GenerateSignedURLs(remotePaths []string, requestType generateSignedURLsRequestType) (*GenerateSignedURLsResponse, error) {
	request := &GenerateSignedURLsRequest{
		Paths: remotePaths,
		Type:  requestType,
	}

	log.Debug("Sending request to generate signed URLs...\n")
	log.Debugf("> Request type: %v\n", requestType)
	log.Debugf("> Paths: %v\n", remotePaths)

	var response *GenerateSignedURLsResponse
	err := retry.RetryWithConstantWait("generate signed URLs", 5, time.Second, func() error {
		r, err := c.executeRequest(request)
		if err != nil {
			return err
		}

		response = r
		return nil
	})

	if err != nil {
		log.Errorf("Error executing signed URLs request: %v\n", err)
		return nil, err
	}

	log.Debugln("Successfully generated signed URLs.\n")
	return response, nil
}

func (c *Client) executeRequest(data interface{}) (*GenerateSignedURLsResponse, error) {
	var b bytes.Buffer
	if data != nil {
		if err := json.NewEncoder(&b).Encode(data); err != nil {
			return nil, fmt.Errorf("failed to encode http data: %v", err)
		}
	}

	q, err := http.NewRequest(http.MethodPost, c.URL, &b)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed URL http request: %v", err)
	}

	q.Header.Set("authorization", c.Token)
	r, err := c.HttpClient.Do(q)
	if err != nil {
		return nil, fmt.Errorf("signed URL request failed: %v", err)
	}

	defer r.Body.Close()

	if !common.IsStatusOK(r.StatusCode) {
		return nil, fmt.Errorf("signed URL request returned %d", r.StatusCode)
	}

	b.Reset()
	tee := io.TeeReader(r.Body, &b)

	var response GenerateSignedURLsResponse
	if err = json.NewDecoder(tee).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode signed URL http response: %v", err)
	}

	if len(response.Error) > 0 {
		return nil, fmt.Errorf("signed URL response returned errors: %s", response.Error)
	}

	return &response, nil
}
