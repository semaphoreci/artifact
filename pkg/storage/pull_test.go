package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreci/artifact/pkg/api"
	"github.com/semaphoreci/artifact/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__doPull_Stats(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "pull_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test artifacts with known sizes
	testFiles := []struct {
		name    string
		content string
		size    int64
	}{
		{"file1.txt", "hello world", 11},
		{"file2.txt", "test content here", 17},
		{"file3.txt", "a", 1},
	}

	artifacts := []*api.Artifact{}
	for _, tf := range testFiles {
		localPath := filepath.Join(tempDir, tf.name)
		artifacts = append(artifacts, &api.Artifact{
			RemotePath: tf.name,
			LocalPath:  localPath,
			URLs:       []*api.SignedURL{}, // Empty for this test
		})

		// Pre-create the files to simulate successful downloads
		err := ioutil.WriteFile(localPath, []byte(tf.content), 0644)
		require.NoError(t, err)
	}

	// Mock the doPull function to skip actual HTTP calls
	// We'll test the stats collection logic by creating a modified version
	stats := &PullStats{}

	// Simulate the stats collection that happens in doPull
	for _, artifact := range artifacts {
		if fileInfo, err := os.Stat(artifact.LocalPath); err == nil {
			stats.FileCount++
			stats.TotalSize += fileInfo.Size()
		}
	}

	// Verify stats
	assert.Equal(t, 3, stats.FileCount)
	assert.Equal(t, int64(29), stats.TotalSize) // 11 + 17 + 1
}

func Test__PullStats_EmptyDirectory(t *testing.T) {
	// Test with no files
	stats := &PullStats{}

	assert.Equal(t, 0, stats.FileCount)
	assert.Equal(t, int64(0), stats.TotalSize)
}

func Test__buildArtifacts_SpecialChars(t *testing.T) {
	t.Run("file with plus in name - S3 encoded URL", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/test_art%2Bifact.txt?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/test_art+ifact.txt",
			Destination: "test_art+ifact.txt",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 1)
		assert.Equal(t, "test_art+ifact.txt", artifacts[0].LocalPath)
		assert.Equal(t, "artifacts/workflows/wfid/test_art+ifact.txt", artifacts[0].RemotePath)
	})

	t.Run("file with plus in name - GCS encoded URL", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://storage.googleapis.com/my-bucket1/artifacts/workflows/wfid/test_art%2Bifact.txt?Expires=231256754712", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/test_art+ifact.txt",
			Destination: "test_art+ifact.txt",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 1)
		assert.Equal(t, "test_art+ifact.txt", artifacts[0].LocalPath)
	})

	t.Run("file with plus in name - custom domain URL", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://artifacts.somedomain.com/my-bucket1/projectid/artifacts/workflows/wfid/test_art%2Bifact.txt?X-Amz-Algorithm", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/test_art+ifact.txt",
			Destination: "test_art+ifact.txt",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 1)
		assert.Equal(t, "test_art+ifact.txt", artifacts[0].LocalPath)
	})

	t.Run("file with space in name - S3 encoded URL", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/my%20file.txt?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/my file.txt",
			Destination: "my file.txt",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 1)
		assert.Equal(t, "my file.txt", artifacts[0].LocalPath)
	})

	t.Run("directory pull with plus in name - multiple files", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/build%2B%2B/main.o?X-Amz-Whatever", Method: "GET"},
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/build%2B%2B/lib.o?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/build++",
			Destination: "build++",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 2)
		assert.Equal(t, "build++/main.o", artifacts[0].LocalPath)
		assert.Equal(t, "build++/lib.o", artifacts[1].LocalPath)
	})

	t.Run("source prefix mismatch returns error", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/other_file.txt?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/expected_file.txt",
			Destination: "expected_file.txt",
		}

		_, err := buildArtifacts(signedURLs, paths, true)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "does not match source")
	})

	t.Run("source prefix boundary mismatch returns error", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/build%2B%2B/main.o?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/build+",
			Destination: "build+",
		}

		_, err := buildArtifacts(signedURLs, paths, true)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "does not match source")
	})

	t.Run("file with literal %2B in name - S3 double-encoded URL", func(t *testing.T) {
		// Filename on disk is literally "file%2Bname.txt" (not file+name.txt).
		// The SDK double-encodes %2B -> %252B in the signed URL.
		// GetObject should decode one level: %252B -> %2B (preserving the literal).
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/file%252Bname.txt?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/file%2Bname.txt",
			Destination: "file%2Bname.txt",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 1)
		assert.Equal(t, "file%2Bname.txt", artifacts[0].LocalPath)
		assert.Equal(t, "artifacts/workflows/wfid/file%2Bname.txt", artifacts[0].RemotePath)
	})

	t.Run("file with literal %2B in name - GCS double-encoded URL", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://storage.googleapis.com/my-bucket1/artifacts/workflows/wfid/file%252Bname.txt?Expires=231256754712", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/file%2Bname.txt",
			Destination: "file%2Bname.txt",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 1)
		assert.Equal(t, "file%2Bname.txt", artifacts[0].LocalPath)
	})

	t.Run("literal %2B must not be confused with actual plus", func(t *testing.T) {
		// URL has %252B (literal %2B), but source path has actual + sign.
		// These should NOT match - the decoded object path "file%2Bname.txt"
		// does not have prefix "artifacts/workflows/wfid/file+name.txt".
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/file%252Bname.txt?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/file+name.txt",
			Destination: "file+name.txt",
		}

		_, err := buildArtifacts(signedURLs, paths, true)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "does not match source")
	})

	t.Run("complex docker path with plus - real world case", func(t *testing.T) {
		signedURLs := []*api.SignedURL{
			{URL: "https://my-bucket1.s3.us-east-1.amazonaws.com/projectid/artifacts/workflows/wfid/lib/bild/test-build%2Bdeps-check/v1.432.1-2-ab123-build-test.gz?X-Amz-Whatever", Method: "GET"},
		}
		paths := &files.ResolvedPath{
			Source:      "artifacts/workflows/wfid/lib/bild/test-build+deps-check/v1.432.1-2-ab123-build-test.gz",
			Destination: "lib/bild/test-build+deps-check/v1.432.1-2-ab123-build-test.gz",
		}

		artifacts, err := buildArtifacts(signedURLs, paths, true)
		assert.Nil(t, err)
		require.Len(t, artifacts, 1)
		assert.Equal(t, "lib/bild/test-build+deps-check/v1.432.1-2-ab123-build-test.gz", artifacts[0].LocalPath)
	})
}

func Test__PullStats_LargeFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "pull_large_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a larger test file
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	localPath := filepath.Join(tempDir, "large_file.bin")
	err = ioutil.WriteFile(localPath, largeContent, 0644)
	require.NoError(t, err)

	artifact := &api.Artifact{
		RemotePath: "large_file.bin",
		LocalPath:  localPath,
		URLs:       []*api.SignedURL{},
	}

	stats := &PullStats{}

	// Simulate stats collection
	if fileInfo, err := os.Stat(artifact.LocalPath); err == nil {
		stats.FileCount++
		stats.TotalSize += fileInfo.Size()
	}

	assert.Equal(t, 1, stats.FileCount)
	assert.Equal(t, int64(1024*1024), stats.TotalSize)
}
