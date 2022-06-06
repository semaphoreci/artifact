package hub

import (
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
