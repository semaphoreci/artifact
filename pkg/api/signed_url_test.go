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

	t.Run("GCS - file with plus", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://storage.googleapis.com/my-bucket1/artifacts/project/projectid/test_art%2Bifact.txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/test_art+ifact.txt", obj)
	})

	t.Run("GCS - file with space", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://storage.googleapis.com/my-bucket1/artifacts/project/projectid/my%20file.txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/my file.txt", obj)
	})

	t.Run("GCS - file with multiple special chars", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://storage.googleapis.com/my-bucket1/artifacts/project/projectid/build%2B%2B/output%20(1).txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/build++/output (1).txt", obj)
	})

	t.Run("GCS - file with percent literal", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://storage.googleapis.com/my-bucket1/artifacts/project/projectid/100%25done.txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/100%done.txt", obj)
	})

	t.Run("GCS - file with literal %2B in name", func(t *testing.T) {
		// Filename is literally "file%2Bname.txt" on disk.
		// SDK double-encodes: %2B -> %252B in the signed URL.
		signedURL := SignedURL{URL: "https://storage.googleapis.com/my-bucket1/artifacts/project/projectid/file%252Bname.txt?Expires=231256754712"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/file%2Bname.txt", obj)
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

	t.Run("S3 - file with plus", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/test_art%2Bifact.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/test_art+ifact.txt", obj)
	})

	t.Run("S3 - file with space", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/my%20file.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/my file.txt", obj)
	})

	t.Run("S3 - file with multiple special chars", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/build%2B%2B/output%20(1).txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/build++/output (1).txt", obj)
	})

	t.Run("S3 - file with percent literal", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/100%25done.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/100%done.txt", obj)
	})

	t.Run("S3 - file with literal %2B in name", func(t *testing.T) {
		// Filename is literally "file%2Bname.txt" on disk.
		// SDK double-encodes: %2B -> %252B in the signed URL.
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/file%252Bname.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/file%2Bname.txt", obj)
	})

	t.Run("S3 region-less - file with plus", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.amazonaws.com/projectid/artifacts/project/projectid/test_art%2Bifact.txt?X-Amz-Whatever"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/test_art+ifact.txt", obj)
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

	t.Run("127.0.0.1 - file with plus", func(t *testing.T) {
		signedURL := SignedURL{URL: "http://127.0.0.1:8080/artifacts/project/projectid/test_art%2Bifact.txt"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/test_art+ifact.txt", obj)
	})

	t.Run("127.0.0.1 - file with space", func(t *testing.T) {
		signedURL := SignedURL{URL: "http://127.0.0.1:8080/artifacts/project/projectid/my%20file.txt"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/my file.txt", obj)
	})

	t.Run("127.0.0.1 - file with literal %2B in name", func(t *testing.T) {
		signedURL := SignedURL{URL: "http://127.0.0.1:8080/artifacts/project/projectid/file%252Bname.txt"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/file%2Bname.txt", obj)
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

	t.Run("custom domain - file with plus", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://artifacts.somedomain.com/my-bucket1/projectid/artifacts/project/projectid/test_art%2Bifact.txt?X-Amz-Algorithm"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/test_art+ifact.txt", obj)
	})

	t.Run("custom domain - file with space", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://artifacts.somedomain.com/my-bucket1/projectid/artifacts/project/projectid/my%20file.txt?X-Amz-Algorithm"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/my file.txt", obj)
	})

	t.Run("custom domain - file with multiple special chars", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://artifacts.somedomain.com/my-bucket1/projectid/artifacts/project/projectid/build%2B%2B/output%20(1).txt?X-Amz-Algorithm"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/build++/output (1).txt", obj)
	})

	t.Run("custom domain - file with literal %2B in name", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://artifacts.somedomain.com/my-bucket1/projectid/artifacts/project/projectid/file%252Bname.txt?X-Amz-Algorithm"}
		obj, err := signedURL.GetObject()
		assert.Nil(t, err)
		assert.Equal(t, "artifacts/project/projectid/file%2Bname.txt", obj)
	})

	t.Run("S3 - malformed URL escape returns parse error", func(t *testing.T) {
		signedURL := SignedURL{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/project/projectid/bad%2Xname.txt?X-Amz-Whatever"}
		_, err := signedURL.GetObject()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to parse URL")
	})

	t.Run("bad URL", func(t *testing.T) {
		signedURL := SignedURL{URL: "http://somehost.com/projectid/artifacts/project/projectid/myfile.txt"}
		_, err := signedURL.GetObject()
		assert.NotNil(t, err)
	})
}
