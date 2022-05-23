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

type pullTestCase struct {
	EnvVar               string
	Prefix               string
	CategoryOverrideFlag string
	Command              func() *cobra.Command
}

func Test__Pull(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	testCases := []pullTestCase{
		{
			EnvVar:               "SEMAPHORE_PROJECT_ID",
			Prefix:               "projects",
			CategoryOverrideFlag: "project-id",
			Command: func() *cobra.Command {
				return NewPullProjectCmd()
			},
		},
		{
			EnvVar:               "SEMAPHORE_WORKFLOW_ID",
			Prefix:               "workflows",
			CategoryOverrideFlag: "workflow-id",
			Command: func() *cobra.Command {
				return NewPullWorkflowCmd()
			},
		},
		{
			EnvVar:               "SEMAPHORE_JOB_ID",
			Prefix:               "jobs",
			CategoryOverrideFlag: "job-id",
			Command: func() *cobra.Command {
				return NewPullJobCmd()
			},
		},
	}

	for _, testCase := range testCases {
		runForPullTestCase(t, testCase)
	}
}

func runForPullTestCase(t *testing.T, testCase pullTestCase) {
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

	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hubServer.URL())
	os.Setenv(testCase.EnvVar, "1")

	t.Run(testCase.Prefix+" missing file", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"notfound.txt"})
		cmd.Execute()

		assertFileDoesNotExist(t, "notfound.txt")
	})

	t.Run(testCase.Prefix+" missing dir", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"notfound/"})
		cmd.Execute()

		assertFileDoesNotExist(t, "notfound")
	})

	t.Run(testCase.Prefix+" single file", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"file1.txt"})
		cmd.Execute()

		assert.FileExists(t, "file1.txt")
		os.Remove("file1.txt")
	})

	t.Run(testCase.Prefix+" single file with destination", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"file1.txt"})
		cmd.Flags().Set("destination", "another.txt")
		cmd.Execute()

		assert.FileExists(t, "another.txt")
		assertFileDoesNotExist(t, "file1.txt")
		os.Remove("another.txt")
	})

	t.Run(testCase.Prefix+" single-level dir", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"one-level/"})
		cmd.Execute()

		assert.DirExists(t, "one-level")
		assert.FileExists(t, "one-level/file1.txt")
		assert.FileExists(t, "one-level/file2.txt")
		os.RemoveAll("one-level")
	})

	t.Run(testCase.Prefix+" single-level dir with destination", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"one-level/"})
		cmd.Flags().Set("destination", "another")
		cmd.Execute()

		assert.DirExists(t, "another")
		assert.FileExists(t, "another/file1.txt")
		assert.FileExists(t, "another/file2.txt")
		os.RemoveAll("another")
	})

	t.Run(testCase.Prefix+" two-levels dir", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"two-levels/"})
		cmd.Execute()

		assert.DirExists(t, "two-levels")
		assert.FileExists(t, "two-levels/file1.txt")
		assert.DirExists(t, "two-levels/sub")
		assert.FileExists(t, "two-levels/sub/file1.txt")
		os.RemoveAll("two-levels")
	})

	t.Run(testCase.Prefix+" overriding category id", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"another.txt"})
		cmd.Flags().Set(testCase.CategoryOverrideFlag, "2")
		cmd.Execute()

		assert.FileExists(t, "another.txt")
		os.Remove("another.txt")
	})
}

func assertFileDoesNotExist(t *testing.T, fileName string) {
	_, err := os.Stat(fileName)
	assert.True(t, os.IsNotExist(err))
}
