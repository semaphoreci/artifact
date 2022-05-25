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

	storage, hub, err := prepare()
	if !assert.Nil(t, err) {
		return
	}

	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hub.URL())
	os.Setenv("SEMAPHORE_JOB_ID", "1")

	t.Run("missing file", func(t *testing.T) {
		output, err := executeCommand("pull", rootFolder, []string{"notfound.txt"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pulling artifact")
		assert.Contains(t, output, "Please check if the artifact you are trying to pull exists")
	})

	t.Run("pulling single file that exists locally throws error", func(t *testing.T) {
		output, err := executeCommand("pull", rootFolder, []string{"file1.txt"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executeCommand("pull", rootFolder, []string{"file1.txt"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pulling artifact")
		assert.Contains(t, output, "'file1.txt' already exists locally; delete it first, or use --force flag")
		os.Remove("file1.txt")
	})

	t.Run("pulling single file that exists locally forcefully works", func(t *testing.T) {
		output, err := executeCommand("pull", rootFolder, []string{"file1.txt"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executeCommand("pull", rootFolder, []string{"file1.txt", "-f"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")
		os.Remove("file1.txt")
	})

	t.Run("pulling directory that exists locally throws error", func(t *testing.T) {
		output, err := executeCommand("pull", rootFolder, []string{"one-level"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executeCommand("pull", rootFolder, []string{"one-level"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pulling artifact")
		assert.Contains(t, output, "'one-level/file1.txt' already exists locally; delete it first, or use --force flag")
		os.RemoveAll("one-level")
	})

	t.Run("pulling directory that has one single file locally throws error", func(t *testing.T) {
		assert.Nil(t, os.Mkdir("one-level", 0755))
		ioutil.WriteFile("one-level/file2.txt", []byte("file2"), 0755)

		output, err := executeCommand("pull", rootFolder, []string{"one-level"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pulling artifact")
		assert.Contains(t, output, "'one-level/file2.txt' already exists locally; delete it first, or use --force flag")
		os.RemoveAll("one-level")
	})

	t.Run("pulling only file from directory that doesn't exist locally works", func(t *testing.T) {
		assert.Nil(t, os.Mkdir("one-level", 0755))
		ioutil.WriteFile("one-level/file2.txt", []byte("file2"), 0755)

		output, err := executeCommand("pull", rootFolder, []string{"one-level/file1.txt"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")
		os.Remove("file1.txt")
		os.RemoveAll("one-level")
	})

	t.Run("pulling directory that exists locally forcefully works", func(t *testing.T) {
		output, err := executeCommand("pull", rootFolder, []string{"one-level"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		output, err = executeCommand("pull", rootFolder, []string{"one-level", "-f"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")
		os.RemoveAll("one-level")
	})

	hub.Close()
	storage.Close()
}

func Test__Push(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	integrationFolder := filepath.Dir(file)
	testFolder := filepath.Dir(integrationFolder)
	rootFolder := filepath.Dir(testFolder)

	storage, hub, err := prepare()
	if !assert.Nil(t, err) {
		return
	}

	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hub.URL())
	os.Setenv("SEMAPHORE_JOB_ID", "1")

	t.Run("pushing empty file", func(t *testing.T) {
		tmpFilePath, _ := ioutil.TempFile("", "*.file")
		tmpFileName := filepath.Base(tmpFilePath.Name())

		output, err := executeCommand("push", rootFolder, []string{tmpFilePath.Name()})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pushed artifact for current job")
		_ = os.Remove(tmpFilePath.Name())

		output, err = executeCommand("pull", rootFolder, []string{tmpFileName})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		fileInfo, err := os.Stat(tmpFileName)
		assert.Nil(t, err)
		assert.Zero(t, fileInfo.Size())
		_ = os.Remove(tmpFileName)
	})

	t.Run("pushing single file that exists remotely throws error", func(t *testing.T) {
		tmpFile, _ := ioutil.TempFile("", "")
		tmpFile.Write([]byte("file1"))

		output, err := executeCommand("push", rootFolder, []string{tmpFile.Name(), "-d", "file1.txt"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pushing artifact")
		assert.Contains(t, output, "'artifacts/jobs/1/file1.txt' already exists in the remote storage; delete it first, or use --force flag")
		os.Remove(tmpFile.Name())
	})

	t.Run("pushing single file that exists remotely forcefully works", func(t *testing.T) {
		tmpFile, _ := ioutil.TempFile("", "")
		tmpFile.Write([]byte("file1"))

		output, err := executeCommand("push", rootFolder, []string{tmpFile.Name(), "-d", "file1.txt", "-f"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pushed artifact for current job")
		os.Remove(tmpFile.Name())
	})

	t.Run("pushing whole directory that exists remotely throws error", func(t *testing.T) {
		tmpDir, _ := ioutil.TempDir("", "")
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file1.txt", tmpDir), []byte("file1"), 0755)
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file2.txt", tmpDir), []byte("file2"), 0755)

		output, err := executeCommand("push", rootFolder, []string{tmpDir, "-d", "one-level"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pushing artifact")
		assert.Contains(t, output, "'artifacts/jobs/1/one-level/file1.txt' already exists in the remote storage; delete it first, or use --force flag")
		os.RemoveAll(tmpDir)
	})

	t.Run("pushing whole directory that exists remotely forcefully works", func(t *testing.T) {
		tmpDir, _ := ioutil.TempDir("", "")
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file1.txt", tmpDir), []byte("file1"), 0755)
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file2.txt", tmpDir), []byte("file2"), 0755)

		output, err := executeCommand("push", rootFolder, []string{tmpDir, "-d", "one-level", "-f"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pushed artifact for current job")
		os.RemoveAll(tmpDir)
	})

	t.Run("pushing directory with one single file that exists remotely throws error", func(t *testing.T) {
		tmpDir, _ := ioutil.TempDir("", "")
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file111.txt", tmpDir), []byte("file111"), 0755)
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file2.txt", tmpDir), []byte("file2"), 0755)

		output, err := executeCommand("push", rootFolder, []string{tmpDir, "-d", "one-level"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error pushing artifact")
		assert.Contains(t, output, "'artifacts/jobs/1/one-level/file2.txt' already exists in the remote storage; delete it first, or use --force flag")
		os.RemoveAll(tmpDir)
	})

	t.Run("pushing directory with one single file that exists remotely forcefully works", func(t *testing.T) {
		tmpDir, _ := ioutil.TempDir("", "")
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file111.txt", tmpDir), []byte("file111"), 0755)
		_ = ioutil.WriteFile(fmt.Sprintf("%s/file2.txt", tmpDir), []byte("file2"), 0755)

		output, err := executeCommand("push", rootFolder, []string{tmpDir, "-d", "one-level", "-f"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pushed artifact for current job")
		os.RemoveAll(tmpDir)
	})

	t.Run("push using input from a pipe", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip()
		}

		command := fmt.Sprintf("echo -n \"hello from pipe\" | %s push job - -d from-pipe.txt -v", getBinaryPath(rootFolder))
		tmpScript, err := createTempScript(command)
		if !assert.Nil(t, err) {
			return
		}

		output, err := executeTempScript(tmpScript)
		assert.Nil(t, err)
		assert.Contains(t, output, "Detected stdin, saving it to a temporary file...")
		assert.Contains(t, output, "Successfully pushed artifact for current job")

		output, err = executeCommand("pull", rootFolder, []string{"from-pipe.txt"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")

		fileContents, _ := ioutil.ReadFile("from-pipe.txt")
		assert.Equal(t, "hello from pipe", string(fileContents))

		os.Remove("from-pipe.txt")
		os.Remove(tmpScript)
	})

	t.Run("push gzipped file from a pipe", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip()
		}

		// Run artifact push
		command := fmt.Sprintf("docker image save alpine/helm | gzip | %s push job - -d docker-image.gz -v", getBinaryPath(rootFolder))
		tmpScript, err := createTempScript(command)
		if !assert.Nil(t, err) {
			return
		}

		output, err := executeTempScript(tmpScript)
		assert.Nil(t, err)
		assert.Contains(t, output, "Detected stdin, saving it to a temporary file...")
		assert.Contains(t, output, "Successfully pushed artifact for current job")

		// Pull uploaded artifact
		output, err = executeCommand("pull", rootFolder, []string{"docker-image.gz"})
		assert.Nil(t, err)
		assert.Contains(t, output, "Successfully pulled artifact for current job")
		assert.FileExists(t, "docker-image.gz")

		// Validate that you can decompress the compressed pulled artifact
		cmd := exec.Command("gzip", "-d", "docker-image.gz")
		_, err = cmd.CombinedOutput()
		assert.Nil(t, err)

		os.Remove("docker-image")
		os.Remove("docker-image.gz")
		os.Remove(tmpScript)
	})

	hub.Close()
	storage.Close()
}

func Test__Yank(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	integrationFolder := filepath.Dir(file)
	testFolder := filepath.Dir(integrationFolder)
	rootFolder := filepath.Dir(testFolder)

	storage, hub, err := prepare()
	if !assert.Nil(t, err) {
		return
	}

	os.Setenv("SEMAPHORE_ARTIFACT_TOKEN", "dummy")
	os.Setenv("SEMAPHORE_ORGANIZATION_URL", hub.URL())
	os.Setenv("SEMAPHORE_JOB_ID", "1")

	t.Run("missing file", func(t *testing.T) {
		output, err := executeCommand("yank", rootFolder, []string{"notfound.txt"})
		assert.NotNil(t, err)
		assert.Contains(t, output, "Error yanking artifact")
		assert.Contains(t, output, "Please check if the artifact you are trying to yank exists")
	})

	hub.Close()
	storage.Close()
}

func prepare() (*testsupport.StorageMockServer, *testsupport.HubMockServer, error) {
	storageServer, err := testsupport.NewStorageMockServer()
	if err != nil {
		return nil, nil, err
	}

	err = storageServer.Init([]testsupport.FileMock{
		{Name: "artifacts/jobs/1/file1.txt", Contents: "something"},
		{Name: "artifacts/jobs/1/file2.txt", Contents: "something"},
		{Name: "artifacts/jobs/1/one-level/file1.txt", Contents: "something"},
		{Name: "artifacts/jobs/1/one-level/file2.txt", Contents: "something"},
		{Name: "artifacts/jobs/1/two-levels/file1.txt", Contents: "something"},
		{Name: "artifacts/jobs/1/two-levels/sub/file1.txt", Contents: "something"},
		{Name: "artifacts/jobs/2/another.txt", Contents: "something"},
	})
	if err != nil {
		storageServer.Close()
		return nil, nil, err
	}

	hubServer := testsupport.NewHubMockServer(storageServer)
	hubServer.Init()

	return storageServer, hubServer, nil
}

func executeCommand(command, rootFolder string, args []string) (string, error) {
	binary := getBinaryPath(rootFolder)
	fullArgs := []string{command, "job"}
	fullArgs = append(fullArgs, args...)
	fullArgs = append(fullArgs, "-v")

	cmd := exec.Command(binary, fullArgs...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

func createTempScript(command string) (string, error) {
	tmpScript, err := ioutil.TempFile("", "*.sh")
	if err != nil {
		return "", err
	}

	defer tmpScript.Close()

	_, err = tmpScript.Write([]byte(command))
	if err != nil {
		return "", err
	}

	return tmpScript.Name(), nil
}

func executeTempScript(tmpScript string) (string, error) {
	cmd := exec.Command("bash", tmpScript)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func getBinaryPath(rootFolder string) string {
	return fmt.Sprintf("%s/artifact", rootFolder)
}
