package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	api "github.com/semaphoreci/artifact/pkg/api"
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

func (c *Client) GenerateSignedURLs(remotePaths []string, requestType GenerateSignedURLsRequestType) (*GenerateSignedURLsResponse, error) {
	req_body := GenerateSignedURLsRequest{
		Paths: remotePaths,
		Type:  requestType,
	}

	log.Debug("Sending request to generate signed URLs...\n")
	log.Debugf("* Request type: %v\n", requestType)
	log.Debugf("* Paths: %v\n", remotePaths)

	var response GenerateSignedURLsResponse

	req, err := createRequest("POST", c.URL, c.Token, req_body)
	if err != nil {
		return nil, err
	}
	retry_client := retryablehttp.NewClient()
	retry_client.RetryMax = 5
	retry_client.RetryWaitMax = 5
	http_resp, err := retry_client.Do(req)
	if err != nil {
		return nil, err
	}
	err = decodeResponse(http_resp, &response)
	if err != nil {
		return nil, err
	}

	log.Debug("Successfully generated signed URLs.\n")
	return &response, nil
}

func createRequest(method, url, token string, req_body interface{}) (*retryablehttp.Request, error) {
	var serialized_request_data bytes.Buffer
	if err := json.NewEncoder(&serialized_request_data).Encode(req_body); err != nil {
		return nil, fmt.Errorf("Failed to encode http data: %v", err)
	}
	req, err := retryablehttp.NewRequest(method, url, serialized_request_data.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Failed to create new Request: %v", err)
	}
	req.Header.Set("authorization", token)
	return req, nil
}

func decodeResponse(http_resp *http.Response, response *GenerateSignedURLsResponse) error {
	defer http_resp.Body.Close()
	if err := json.NewDecoder(http_resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode signed URL http response: %v", err)
	}

	if len(response.Error) > 0 {
		return fmt.Errorf("signed URL response returned errors: %s", response.Error)
	}

	return nil
}
