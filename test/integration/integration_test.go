package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	testsupport "github.com/semaphoreci/artifact/test/support"
	"github.com/stretchr/testify/assert"
)

func Test__Pull(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	integrationFolder := filepath.Dir(file)
	testFolder := filepath.Dir(integrationFolder)
	rootFolder := filepath.Dir(testFolder)

	storage, hub := prepare()
	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hub.URL())
	os.Setenv("SEMAPHORE_JOB_ID", "1")

	t.Run("pulling single file that exists locally throws error", func(t *testing.T) {
		output, err := executePull(rootFolder, []string{"file1.txt"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executePull(rootFolder, []string{"file1.txt"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pulling artifact")
		assert.Contains(t, output, "'file1.txt' already exists locally; delete it first, or use --force flag")
		os.Remove("file1.txt")
	})

	t.Run("pulling single file that exists locally forcefully works", func(t *testing.T) {
		output, err := executePull(rootFolder, []string{"file1.txt"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executePull(rootFolder, []string{"file1.txt", "-f"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")
		os.Remove("file1.txt")
	})

	t.Run("pulling directory that exists locally throws error", func(t *testing.T) {
		output, err := executePull(rootFolder, []string{"one-level"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executePull(rootFolder, []string{"one-level"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pulling artifact")
		assert.Contains(t, output, "'one-level/file1.txt' already exists locally; delete it first, or use --force flag")
		os.RemoveAll("one-level")
	})

	t.Run("pulling directory that has one single file locally throws error", func(t *testing.T) {
		assert.Nil(t, os.Mkdir("one-level", 0755))
		ioutil.WriteFile("one-level/file2.txt", []byte("file2"), 0755)

		output, err := executePull(rootFolder, []string{"one-level"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pulling artifact")
		assert.Contains(t, output, "'one-level/file2.txt' already exists locally; delete it first, or use --force flag")
		os.RemoveAll("one-level")
	})

	t.Run("pulling only file from directory that doesn't exist locally works", func(t *testing.T) {
		assert.Nil(t, os.Mkdir("one-level", 0755))
		ioutil.WriteFile("one-level/file2.txt", []byte("file2"), 0755)

		output, err := executePull(rootFolder, []string{"one-level/file1.txt"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")
		os.Remove("file1.txt")
		os.RemoveAll("one-level")
	})

	t.Run("pulling directory that exists locally forcefully works", func(t *testing.T) {
		output, err := executePull(rootFolder, []string{"one-level"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executePull(rootFolder, []string{"one-level", "-f"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")
		os.RemoveAll("one-level")
	})

	hub.Close()
	storage.Close()
}

func prepare() (*testsupport.StorageMockServer, *testsupport.HubMockServer) {
	storageServer := testsupport.NewStorageMockServer()
	storageServer.Init([]string{
		"artifacts/jobs/1/file1.txt",
		"artifacts/jobs/1/file2.txt",
		"artifacts/jobs/1/one-level/file1.txt",
		"artifacts/jobs/1/one-level/file2.txt",
		"artifacts/jobs/1/two-levels/file1.txt",
		"artifacts/jobs/1/two-levels/sub/file1.txt",
		"artifacts/jobs/2/another.txt",
	})

	hubServer := testsupport.NewHubMockServer(storageServer)
	hubServer.Init()

	return storageServer, hubServer
}

func executePull(rootFolder string, args []string) (string, error) {
	binary := getBinaryPath(rootFolder)

	fullArgs := []string{"pull", "job"}
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command(binary, fullArgs...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

func getBinaryPath(rootFolder string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s/artifact.exe", rootFolder)
	}

	return fmt.Sprintf("%s/artifact", rootFolder)
}
