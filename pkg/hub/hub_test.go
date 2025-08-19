package hub

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__HubClientCreation(t *testing.T) {
	t.Run("SEMAPHORE_ARTIFACT_TOKEN is required", func(t *testing.T) {
		os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "")
		os.Setenv("SEMAPHORE_ORGANIZATION_URL", "http://some-org.com")
		_, err := NewClient()
		if assert.NotNil(t, err) {
			assert.Equal(t, "SEMAPHORE_ARTIFACT_TOKEN is not set", err.Error())
		}
	})

	t.Run("SEMAPHORE_ORGANIZATION_URL is required", func(t *testing.T) {
		os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
		os.Setenv("SEMAPHORE_ORGANIZATION_URL", "")
		_, err := NewClient()
		if assert.NotNil(t, err) {
			assert.Equal(t, "SEMAPHORE_ORGANIZATION_URL is not set", err.Error())
		}
	})

	t.Run("bad SEMAPHORE_ORGANIZATION_URL throws error", func(t *testing.T) {
		os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
		os.Setenv("SEMAPHORE_ORGANIZATION_URL", ":asdasd")
		_, err := NewClient()
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "failed to parse SEMAPHORE_ORGANIZATION_URL")
		}
	})

	t.Run("client is created if all parameters are ok", func(t *testing.T) {
		os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
		os.Setenv("SEMAPHORE_ORGANIZATION_URL", "https://myorg.semaphoreci.com")
		client, err := NewClient()
		assert.Nil(t, err)
		assert.NotNil(t, client)
	})
}

func Test__GenerateSignedURL(t *testing.T) {

	t.Run("response has invalid JSON", func(t *testing.T) {
		noOfCalls := 0
		mockArtifactHubServer := generateMockServer(&noOfCalls, 200, []byte(""))
		defer mockArtifactHubServer.Close()

		response, err := generateSignedURLsHelper(mockArtifactHubServer.URL)
		assert.Nil(t, response)
		assert.Equal(t, 1, noOfCalls)
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "failed to decode signed URL http response")
		}
	})

	t.Run("Retry only once when artifact hub returns 404", func(t *testing.T) {
		noOfCalls := 0
		mockArtifactHubServer := generateMockServer(&noOfCalls, 404, []byte("{}"))
		defer mockArtifactHubServer.Close()

		response, err := generateSignedURLsHelper(mockArtifactHubServer.URL)
		assert.Nil(t, response)
		assert.Equal(t, 1, noOfCalls)
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "hub returned 404 status code")
		}
	})

	t.Run("Retry only once when artifact hub returns 401", func(t *testing.T) {
		noOfCalls := 0
		mockArtifactHubServer := generateMockServer(&noOfCalls, 401, []byte("{}"))
		defer mockArtifactHubServer.Close()

		response, err := generateSignedURLsHelper(mockArtifactHubServer.URL)
		assert.Nil(t, response)
		assert.Equal(t, 1, noOfCalls)
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "hub returned 401 status code")
		}
	})

	t.Run("Retry 5 times when artifact hub returns 500", func(t *testing.T) {
		noOfCalls := 0
		mockArtifactHubServer := generateMockServer(&noOfCalls, 500, []byte("{}"))
		defer mockArtifactHubServer.Close()

		response, err := generateSignedURLsHelper(mockArtifactHubServer.URL)
		assert.Nil(t, response)
		assert.Equal(t, 5, noOfCalls)
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "request did not return a non-5xx response")
		}
	})

	t.Run("Check if http_proxy enviroment variable is used", func(t *testing.T) {
		noOfCalls := 0
		mockArtifactHubServer := generateMockServer(&noOfCalls, 500, []byte("{}"))
		defer mockArtifactHubServer.Close()

		generateSignedURLsHelper(mockArtifactHubServer.URL)

	})
}

func generateSignedURLsHelper(url string) (*GenerateSignedURLsResponse, error) {
	client := Client{
		URL:        url,
		Token:      "",
		HttpClient: &http.Client{},
	}
	return client.GenerateSignedURLs([]string{}, GenerateSignedURLsRequestPULL)
}

func generateMockServer(counter *int, codeToReturn int, responseBody []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*counter++
		w.WriteHeader(codeToReturn)
		w.Write(responseBody)
	}))
}
