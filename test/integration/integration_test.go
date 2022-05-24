package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	testsupport "github.com/semaphoreci/artifact/test/support"
	"github.com/stretchr/testify/assert"
)

func Test__PullForcifully(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	integrationFolder := filepath.Dir(file)
	testFolder := filepath.Dir(integrationFolder)
	rootFolder := filepath.Dir(testFolder)

	storage, hub := prepare()
	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hub.URL())
	os.Setenv("SEMAPHORE_JOB_ID", "1")

	_, err := executePull(rootFolder, "file1.txt")
	assert.Nil(t, err)
	// TODO: assert output

	_, err = executePull(rootFolder, "file1.txt")
	assert.NotNil(t, err)
	// TODO: assert output

	hub.Close()
	storage.Close()
	os.Remove("file1.txt")
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

func executePull(rootFolder, fileName string) (string, error) {
	binary := getBinaryPath(rootFolder)
	cmd := exec.Command(binary, "pull", "job", fileName)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func getBinaryPath(rootFolder string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s/artifact.exe", rootFolder)
	}

	return fmt.Sprintf("%s/artifact", rootFolder)
}
