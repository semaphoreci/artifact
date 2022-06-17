package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	testsupport "github.com/semaphoreci/artifact/test/support"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type pushTestCase struct {
	EnvVar               string
	Prefix               string
	CategoryOverrideFlag string
	Command              func() *cobra.Command
}

func Test__Push(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	testCases := []pushTestCase{
		{
			EnvVar:               "SEMAPHORE_PROJECT_ID",
			Prefix:               "projects",
			CategoryOverrideFlag: "project-id",
			Command: func() *cobra.Command {
				return NewPushProjectCmd()
			},
		},
		{
			EnvVar:               "SEMAPHORE_WORKFLOW_ID",
			Prefix:               "workflows",
			CategoryOverrideFlag: "workflow-id",
			Command: func() *cobra.Command {
				return NewPushWorkflowCmd()
			},
		},
		{
			EnvVar:               "SEMAPHORE_JOB_ID",
			Prefix:               "jobs",
			CategoryOverrideFlag: "job-id",
			Command: func() *cobra.Command {
				return NewPushJobCmd()
			},
		},
	}

	for _, testCase := range testCases {
		storageServer, err := testsupport.NewStorageMockServer()
		if !assert.Nil(t, err) {
			return
		}

		storageServer.Init([]testsupport.FileMock{})

		hubServer := testsupport.NewHubMockServer(storageServer)
		hubServer.Init()
		runPushTestCase(t, testCase, hubServer, storageServer)
		hubServer.Close()
		storageServer.Close()
	}
}

func runPushTestCase(t *testing.T, testCase pushTestCase, hub *testsupport.HubMockServer, storage *testsupport.StorageMockServer) {
	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hub.URL())
	os.Setenv(testCase.EnvVar, "1")

	t.Run(testCase.Prefix+" missing file", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"notfound.txt"})
		cmd.Execute()

		assert.False(t, storage.IsFile(fmt.Sprintf("artifacts/%s/1/notfound.txt", testCase.Prefix)))
	})

	t.Run(testCase.Prefix+" missing dir", func(t *testing.T) {
		cmd := testCase.Command()
		cmd.SetArgs([]string{"notfound/"})
		cmd.Execute()

		assert.False(t, storage.IsDir(fmt.Sprintf("artifacts/%s/1/notfound/", testCase.Prefix)))
	})

	t.Run(testCase.Prefix+" single file", func(t *testing.T) {
		tempFile, _ := ioutil.TempFile("", "*")
		tempFile.Write([]byte("something"))

		cmd := testCase.Command()
		cmd.SetArgs([]string{tempFile.Name()})
		cmd.Execute()

		assert.True(t, storage.IsFile(fmt.Sprintf("artifacts/%s/1/%s", testCase.Prefix, filepath.Base(tempFile.Name()))))
		os.Remove(tempFile.Name())
	})

	t.Run(testCase.Prefix+" single file with destination", func(t *testing.T) {
		tempFile, _ := ioutil.TempFile("", "*")
		tempFile.Write([]byte("something"))

		cmd := testCase.Command()
		cmd.SetArgs([]string{tempFile.Name()})
		cmd.Flags().Set("destination", "myfile")
		cmd.Execute()

		assert.True(t, storage.IsFile(fmt.Sprintf("artifacts/%s/1/myfile", testCase.Prefix)))
		os.Remove(tempFile.Name())
	})

	t.Run(testCase.Prefix+" single-level dir", func(t *testing.T) {
		tempDir, _ := ioutil.TempDir("", "*")
		tempFile1, _ := ioutil.TempFile(tempDir, "*")
		tempFile2, _ := ioutil.TempFile(tempDir, "*")
		tempFile1.Write([]byte("something"))
		tempFile2.Write([]byte("something"))

		cmd := testCase.Command()
		cmd.SetArgs([]string{tempDir})
		cmd.Execute()

		assert.True(t, storage.IsDir(
			fmt.Sprintf("artifacts/%s/1/%s",
				testCase.Prefix,
				filepath.Base(tempDir)),
		))

		assert.True(t, storage.IsFile(
			fmt.Sprintf("artifacts/%s/1/%s/%s",
				testCase.Prefix,
				filepath.Base(tempDir),
				filepath.Base(tempFile1.Name())),
		))

		assert.True(t, storage.IsFile(
			fmt.Sprintf("artifacts/%s/1/%s/%s",
				testCase.Prefix,
				filepath.Base(tempDir),
				filepath.Base(tempFile2.Name())),
		))

		os.RemoveAll(tempDir)
	})

	t.Run(testCase.Prefix+" single-level dir with destination", func(t *testing.T) {
		tempDir, _ := ioutil.TempDir("", "*")
		tempFile1, _ := ioutil.TempFile(tempDir, "*")
		tempFile2, _ := ioutil.TempFile(tempDir, "*")
		tempFile1.Write([]byte("something"))
		tempFile2.Write([]byte("something"))

		cmd := testCase.Command()
		cmd.SetArgs([]string{tempDir})
		cmd.Flags().Set("destination", "mydir")
		cmd.Execute()

		assert.True(t, storage.IsDir(fmt.Sprintf("artifacts/%s/1/mydir", testCase.Prefix)))

		assert.True(t, storage.IsFile(
			fmt.Sprintf("artifacts/%s/1/mydir/%s",
				testCase.Prefix,
				filepath.Base(tempFile1.Name())),
		))

		assert.True(t, storage.IsFile(
			fmt.Sprintf("artifacts/%s/1/mydir/%s",
				testCase.Prefix,
				filepath.Base(tempFile2.Name())),
		))

		os.RemoveAll(tempDir)
	})

	t.Run(testCase.Prefix+" two-levels dir", func(t *testing.T) {
		tempDir, _ := ioutil.TempDir("", "*")
		subDir, _ := ioutil.TempDir(tempDir, "*")
		tempFile1, _ := ioutil.TempFile(tempDir, "*.file")
		tempFile2, _ := ioutil.TempFile(subDir, "*.file")
		tempFile1.Write([]byte("something"))
		tempFile2.Write([]byte("something"))

		cmd := testCase.Command()
		cmd.SetArgs([]string{tempDir})
		cmd.Execute()

		assert.True(t, storage.IsDir(
			fmt.Sprintf("artifacts/%s/1/%s",
				testCase.Prefix,
				filepath.Base(tempDir)),
		))

		assert.True(t, storage.IsDir(
			fmt.Sprintf("artifacts/%s/1/%s/%s",
				testCase.Prefix,
				filepath.Base(tempDir),
				filepath.Base(subDir)),
		))

		assert.True(t, storage.IsFile(
			fmt.Sprintf("artifacts/%s/1/%s/%s",
				testCase.Prefix,
				filepath.Base(tempDir),
				filepath.Base(tempFile1.Name())),
		))

		assert.True(t, storage.IsFile(
			fmt.Sprintf("artifacts/%s/1/%s/%s/%s",
				testCase.Prefix,
				filepath.Base(tempDir),
				filepath.Base(subDir),
				filepath.Base(tempFile2.Name())),
		))

		os.RemoveAll(tempDir)
	})

	t.Run(testCase.Prefix+" two-levels dir sub-directory", func(t *testing.T) {
		tempDir, _ := ioutil.TempDir("", "*")
		subDir, _ := ioutil.TempDir(tempDir, "*")
		tempFile1, _ := ioutil.TempFile(tempDir, "*.file")
		tempFile2, _ := ioutil.TempFile(subDir, "*.file")
		tempFile1.Write([]byte("something"))
		tempFile2.Write([]byte("something"))

		cmd := testCase.Command()
		cmd.SetArgs([]string{subDir})
		cmd.Execute()

		assert.False(t, storage.IsDir(
			fmt.Sprintf("artifacts/%s/1/%s",
				testCase.Prefix,
				filepath.Base(tempDir)),
		))

		assert.True(t, storage.IsDir(
			fmt.Sprintf("artifacts/%s/1/%s",
				testCase.Prefix,
				filepath.Base(subDir)),
		))

		assert.True(t, storage.IsFile(
			fmt.Sprintf("artifacts/%s/1/%s/%s",
				testCase.Prefix,
				filepath.Base(subDir),
				filepath.Base(tempFile2.Name())),
		))

		os.RemoveAll(tempDir)
	})

	t.Run(testCase.Prefix+" overriding category id", func(t *testing.T) {
		tempFile, _ := ioutil.TempFile("", "*")
		tempFile.Write([]byte("something"))

		cmd := testCase.Command()
		cmd.SetArgs([]string{tempFile.Name()})
		cmd.Flags().Set(testCase.CategoryOverrideFlag, "2")
		cmd.Execute()

		assert.True(t, storage.IsFile(fmt.Sprintf("artifacts/%s/2/%s", testCase.Prefix, filepath.Base(tempFile.Name()))))
		os.Remove(tempFile.Name())
	})
}
