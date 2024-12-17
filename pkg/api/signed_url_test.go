package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__GetObject(t *testing.T) {
	t.Run("GCS - file", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://storage.googleapis.com/my-bucket1/artifacts/project/projectid/myfile.txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/myfile.txt", obj)
	})

	t.Run("GCS - file inside directory", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://storage.googleapis.com/my-bucket1/artifacts/project/projectid/mydir/myfile.txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/mydir/myfile.txt", obj)
	})

	t.Run("S3 - file", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/myfile.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/myfile.txt", obj)
	})

	t.Run("S3 with region-less URL - file", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.amazonaws.com/projectid/artifacts/project/projectid/myfile.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/myfile.txt", obj)
	})

	t.Run("S3 - file inside directory", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/mydir/myfile.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/mydir/myfile.txt", obj)
	})

	t.Run("S3 with region-less URL - file inside directory", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.amazonaws.com/projectid/artifacts/project/projectid/mydir/myfile.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/mydir/myfile.txt", obj)
	})

	t.Run("127.0.0.1 - file", func(t *testing.T) {
		signedURL := SignedURL{URL: "http://127.0.0.1:8080/artifacts/project/projectid/myfile.txt"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/myfile.txt", obj)
	})

	t.Run("127.0.0.1 - file inside directory", func(t *testing.T) {
		signedURL := SignedURL{URL: "http://127.0.0.1:8080/artifacts/project/projectid/mydir/myfile.txt"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/mydir/myfile.txt", obj)
	})

	t.Run("custom domain - file", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://artifacts.somedomain.com/my-bucket1/projectid/artifacts/project/projectid/myfile.txt?X-Amz-Algorithm"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/myfile.txt", obj)
	})

	t.Run("custom domain - file inside directory", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://artifacts.somedomain.com/my-bucket1/projectid/artifacts/project/projectid/mydir/myfile.txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/mydir/myfile.txt", obj)
	})

	t.Run("bad URL", func(t *testing.T) {
		signedURL := SignedURL{URL: "http://somehost.com/projectid/artifacts/project/projectid/myfile.txt"}
		_, err := signedURL.GetObject()
		assert.NotNil(t, err)
	})
}
