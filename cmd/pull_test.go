package cmd

import (
	"os"
	"testing"

	testsupport "github.com/semaphoreci/artifact/test/support"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test__Pull(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	storageServer := testsupport.NewStorageMockServer()
	storageServer.Init([]string{
		"artifacts/projects/1/file1.txt",
		"artifacts/projects/1/file2.txt",
		"artifacts/projects/1/first/file1.txt",
		"artifacts/projects/1/first/file2.txt",
		"artifacts/projects/1/second/file1.txt",
		"artifacts/workflows/1/file1.txt",
		"artifacts/workflows/1/file2.txt",
		"artifacts/workflows/1/first/file1.txt",
		"artifacts/workflows/1/first/file2.txt",
		"artifacts/workflows/1/second/file1.txt",
		"artifacts/jobs/1/file1.txt",
		"artifacts/jobs/1/file2.txt",
		"artifacts/jobs/1/first/file1.txt",
		"artifacts/jobs/1/first/file2.txt",
		"artifacts/jobs/1/second/file1.txt",
	})

	hubServer := testsupport.NewHubMockServer(storageServer)
	hubServer.Init()

	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hubServer.URL())
	os.Setenv("SEMAPHORE_PROJECT_ID", "1")
	os.Setenv("SEMAPHORE_WORKFLOW_ID", "1")
	os.Setenv("SEMAPHORE_JOB_ID", "1")

	t.Run("existing single file", func(t *testing.T) {
		cmd := NewPullProjectCmd()
		cmd.SetArgs([]string{"file1.txt"})
		cmd.Execute()

		assert.FileExists(t, "file1.txt")
		os.Remove("file1.txt")
	})

	t.Run("missing single file", func(t *testing.T) {
		cmd := NewPullProjectCmd()
		cmd.SetArgs([]string{"notfound.txt"})
		cmd.Execute()

		assertFileDoesNotExist(t, "notfound.txt")
	})

	t.Run("existing dir", func(t *testing.T) {
		cmd := NewPullProjectCmd()
		cmd.SetArgs([]string{"first/"})
		cmd.Execute()

		assert.DirExists(t, "first")
		assert.FileExists(t, "first/file1.txt")
		assert.FileExists(t, "first/file2.txt")
		os.RemoveAll("first")
	})

	t.Run("missing dir", func(t *testing.T) {
		cmd := NewPullProjectCmd()
		cmd.SetArgs([]string{"notfound/"})
		cmd.Execute()

		assertFileDoesNotExist(t, "first/file1.txt")
		assertFileDoesNotExist(t, "first/file2.txt")
	})
}

func assertFileDoesNotExist(t *testing.T, fileName string) {
	_, err := os.Stat(fileName)
	assert.True(t, os.IsNotExist(err))
}
