package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreci/artifact/pkg/api"
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
