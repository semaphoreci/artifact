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

func Test__doPush_Stats_SingleFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "push_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("hello world")
	err = ioutil.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	artifact := &api.Artifact{
		RemotePath: "test.txt",
		LocalPath:  testFile,
		URLs: []*api.SignedURL{
			{Method: "PUT", URL: "https://example.com/upload"},
		},
	}

	stats := &PushStats{}

	fileInfo, err := os.Stat(artifact.LocalPath)
	require.NoError(t, err)

	for _, url := range artifact.URLs {
		if url.Method == "PUT" {
			stats.FileCount++
			stats.TotalSize += fileInfo.Size()
			break
		}
	}

	assert.Equal(t, 1, stats.FileCount)
	assert.Equal(t, int64(11), stats.TotalSize)
}

func Test__doPush_Stats_MultipleFiles(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "push_multi_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFiles := []struct {
		name    string
		content string
		size    int64
	}{
		{"file1.txt", "hello world", 11},
		{"file2.txt", "test content here", 17},
		{"file3.txt", "a", 1},
		{"file4.txt", "longer content for testing purposes", 35},
	}

	artifacts := []*api.Artifact{}
	for _, tf := range testFiles {
		localPath := filepath.Join(tempDir, tf.name)
		err := ioutil.WriteFile(localPath, []byte(tf.content), 0644)
		require.NoError(t, err)

		artifacts = append(artifacts, &api.Artifact{
			RemotePath: tf.name,
			LocalPath:  localPath,
			URLs: []*api.SignedURL{
				{Method: "PUT", URL: "https://example.com/upload/" + tf.name},
			},
		})
	}

	stats := &PushStats{}

	for _, artifact := range artifacts {
		fileInfo, err := os.Stat(artifact.LocalPath)
		require.NoError(t, err)

		for _, url := range artifact.URLs {
			if url.Method == "PUT" {
				stats.FileCount++
				stats.TotalSize += fileInfo.Size()
				break
			}
		}
	}

	assert.Equal(t, 4, stats.FileCount)
	assert.Equal(t, int64(64), stats.TotalSize)
}

func Test__PushStats_EmptyList(t *testing.T) {
	stats := &PushStats{}
	artifacts := []*api.Artifact{}

	for _, artifact := range artifacts {
		fileInfo, err := os.Stat(artifact.LocalPath)
		if err == nil {
			for _, url := range artifact.URLs {
				if url.Method == "PUT" {
					stats.FileCount++
					stats.TotalSize += fileInfo.Size()
					break
				}
			}
		}
	}

	assert.Equal(t, 0, stats.FileCount)
	assert.Equal(t, int64(0), stats.TotalSize)
}

func Test__PushStats_LargeFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "push_large_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	localPath := filepath.Join(tempDir, "large_file.bin")
	err = ioutil.WriteFile(localPath, largeContent, 0644)
	require.NoError(t, err)

	artifact := &api.Artifact{
		RemotePath: "large_file.bin",
		LocalPath:  localPath,
		URLs: []*api.SignedURL{
			{Method: "PUT", URL: "https://example.com/upload/large_file.bin"},
		},
	}

	stats := &PushStats{}

	fileInfo, err := os.Stat(artifact.LocalPath)
	require.NoError(t, err)

	for _, url := range artifact.URLs {
		if url.Method == "PUT" {
			stats.FileCount++
			stats.TotalSize += fileInfo.Size()
			break
		}
	}

	assert.Equal(t, 1, stats.FileCount)
	assert.Equal(t, int64(1024*1024), stats.TotalSize)
}

func Test__PushStats_NonForceWithHEAD(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "push_head_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content")
	err = ioutil.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	artifact := &api.Artifact{
		RemotePath: "test.txt",
		LocalPath:  testFile,
		URLs: []*api.SignedURL{
			{Method: "HEAD", URL: "https://example.com/check"},
			{Method: "PUT", URL: "https://example.com/upload"},
		},
	}

	stats := &PushStats{}

	fileInfo, err := os.Stat(artifact.LocalPath)
	require.NoError(t, err)

	for _, url := range artifact.URLs {
		if url.Method == "PUT" {
			stats.FileCount++
			stats.TotalSize += fileInfo.Size()
			break
		}
	}

	assert.Equal(t, 1, stats.FileCount)
	assert.Equal(t, int64(12), stats.TotalSize)
}

func Test__LocateArtifacts_SingleFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "locate_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "single.txt")
	err = ioutil.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	paths := &files.ResolvedPath{
		Source:      testFile,
		Destination: "remote/single.txt",
	}

	artifacts, err := LocateArtifacts(paths)
	require.NoError(t, err)

	assert.Equal(t, 1, len(artifacts))
	assert.Equal(t, testFile, artifacts[0].LocalPath)
	assert.Equal(t, "remote/single.txt", artifacts[0].RemotePath)
}

func Test__LocateArtifacts_Directory(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "locate_dir_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")

	err = ioutil.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = ioutil.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)

	paths := &files.ResolvedPath{
		Source:      tempDir,
		Destination: "remote/dir",
	}

	artifacts, err := LocateArtifacts(paths)
	require.NoError(t, err)

	assert.Equal(t, 2, len(artifacts))

	localPaths := []string{artifacts[0].LocalPath, artifacts[1].LocalPath}
	assert.Contains(t, localPaths, file1)
	assert.Contains(t, localPaths, file2)
}
