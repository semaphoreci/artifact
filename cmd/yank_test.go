package cmd

import (
	"fmt"
	"os"
	"testing"

	testsupport "github.com/semaphoreci/artifact/test/support"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type yankTestCase struct {
	EnvVar               string
	Prefix               string
	CategoryOverrideFlag string
	Command              func() *cobra.Command
}

func Test__Yank(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	testCases := []yankTestCase{
		{
			EnvVar:               "SEMAPHORE_PROJECT_ID",
			Prefix:               "projects",
			CategoryOverrideFlag: "project-id",
			Command: func() *cobra.Command {
				return NewYankProjectCmd()
			},
		},
		{
			EnvVar:               "SEMAPHORE_WORKFLOW_ID",
			Prefix:               "workflows",
			CategoryOverrideFlag: "workflow-id",
			Command: func() *cobra.Command {
				return NewYankWorkflowCmd()
			},
		},
		{
			EnvVar:               "SEMAPHORE_JOB_ID",
			Prefix:               "jobs",
			CategoryOverrideFlag: "job-id",
			Command: func() *cobra.Command {
				return NewYankJobCmd()
			},
		},
	}

	for _, testCase := range testCases {
		storageServer := testsupport.NewStorageMockServer()
		storageServer.Init([]string{
			fmt.Sprintf("artifacts/%s/1/file1.txt", testCase.Prefix),
			fmt.Sprintf("artifacts/%s/1/file2.txt", testCase.Prefix),
			fmt.Sprintf("artifacts/%s/1/one-level/file1.txt", testCase.Prefix),
			fmt.Sprintf("artifacts/%s/1/one-level/file2.txt", testCase.Prefix),
			fmt.Sprintf("artifacts/%s/1/two-levels/file1.txt", testCase.Prefix),
			fmt.Sprintf("artifacts/%s/1/two-levels/sub/file1.txt", testCase.Prefix),
			fmt.Sprintf("artifacts/%s/2/another.txt", testCase.Prefix),
		})

		hubServer := testsupport.NewHubMockServer(storageServer)
		hubServer.Init()
		runYankTestCase(t, testCase, hubServer, storageServer)
		hubServer.Close()
		storageServer.Close()
	}
}

func runYankTestCase(t *testing.T, testCase yankTestCase, hub *testsupport.HubMockServer, storage *testsupport.StorageMockServer) {
	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hub.URL())
	os.Setenv(testCase.EnvVar, "1")

	t.Run(testCase.Prefix+" single file", func(t *testing.T) {
		fileName := fmt.Sprintf("artifacts/%s/1/file1.txt", testCase.Prefix)
		assert.True(t, storage.IsFile(fileName))

		cmd := testCase.Command()
		cmd.SetArgs([]string{"file1.txt"})
		cmd.Execute()

		assert.False(t, storage.IsFile(fileName))
	})

	t.Run(testCase.Prefix+" single-level dir", func(t *testing.T) {
		dirName := fmt.Sprintf("artifacts/%s/1/one-level/", testCase.Prefix)
		assert.True(t, storage.IsDir(dirName))
		assert.True(t, storage.IsFile(fmt.Sprintf("%sfile1.txt", dirName)))
		assert.True(t, storage.IsFile(fmt.Sprintf("%sfile2.txt", dirName)))

		cmd := testCase.Command()
		cmd.SetArgs([]string{"one-level/"})
		cmd.Execute()

		assert.False(t, storage.IsDir(dirName))
		assert.False(t, storage.IsFile(fmt.Sprintf("%sfile1.txt", dirName)))
		assert.False(t, storage.IsFile(fmt.Sprintf("%sfile2.txt", dirName)))
	})

	t.Run(testCase.Prefix+" two-levels dir", func(t *testing.T) {
		dirName := fmt.Sprintf("artifacts/%s/1/two-levels/", testCase.Prefix)
		subDirName := fmt.Sprintf("artifacts/%s/1/two-levels/sub/", testCase.Prefix)
		assert.True(t, storage.IsDir(dirName))
		assert.True(t, storage.IsDir(subDirName))
		assert.True(t, storage.IsFile(fmt.Sprintf("%sfile1.txt", dirName)))
		assert.True(t, storage.IsFile(fmt.Sprintf("%sfile1.txt", subDirName)))

		cmd := testCase.Command()
		cmd.SetArgs([]string{"two-levels/"})
		cmd.Execute()

		assert.False(t, storage.IsDir(dirName))
		assert.False(t, storage.IsDir(subDirName))
		assert.False(t, storage.IsFile(fmt.Sprintf("%sfile1.txt", dirName)))
		assert.False(t, storage.IsFile(fmt.Sprintf("%sfile1.txt", subDirName)))
	})

	t.Run(testCase.Prefix+" overriding category id", func(t *testing.T) {
		fileName := fmt.Sprintf("artifacts/%s/2/another.txt", testCase.Prefix)
		assert.True(t, storage.IsFile(fileName))

		cmd := testCase.Command()
		cmd.SetArgs([]string{"another.txt"})
		cmd.Flags().Set(testCase.CategoryOverrideFlag, "2")
		cmd.Execute()

		assert.False(t, storage.IsFile(fileName))
	})
}
