package cmd

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"

	"path"
	"testing"
	"time"

	"github.com/semaphoreci/artifact/cmd/utils"
)

var (
	content  = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	content2 = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam lacus massa, porttitor non euismod vel, volutpat eget metus. Maecenas finibus interdum ante id rhoncus. Mauris sodales congue volutpat. Integer scelerisque elit nec lectus varius luctus. Ut tempor orci at tellus facilisis interdum. In scelerisque nec sem vitae euismod. Suspendisse nibh nulla, egestas varius tortor quis, hendrerit cursus urna. Maecenas eu risus ligula. Sed eu tortor orci. Donec mattis cursus gravida.")
)

func assertNilError(t *testing.T, msg string, args ...interface{}) {
	if args[len(args)-1] != nil {
		t.Fatalf(msg+" should be nil, but it's: %s; stack: %s", append(args, string(debug.Stack()))...)
	}
}

func assertTrue(t *testing.T, ok bool, msg string, args ...interface{}) {
	if !ok {
		t.Fatalf(msg+"; stack: %s", append(args, string(debug.Stack()))...)
	}
}

func assertAlreadyExists(t *testing.T, msg string, err error) {
	_, ok := err.(*ErrAlreadyExists)
	assertTrue(t, ok, msg+" should fail with an ErrAlreadyExists, but it's: %s", err)
}

func TestPushPaths(t *testing.T) {
	testPushPaths := func(category, dst, src, expDst, expSrc string) {
		resultDst, resultSrc := pushPaths(category, dst, src)
		if resultDst != expDst {
			t.Errorf("not match destination(%s) with expected(%s) for category(%s), dst(%s) and src(%s)",
				resultDst, expDst, category, dst, src)
		}
		if resultSrc != expSrc {
			t.Errorf("not match source(%s) with expected(%s) for category(%s), dst(%s) and src(%s)",
				resultSrc, expSrc, category, dst, src)
		}
	}

	jobID := "JOB_03"
	os.Setenv(utils.CategoryEnv[utils.JOB], jobID)
	testPushPaths(utils.JOB, "", "x.zip", "/artifacts/jobs/"+jobID+"/x.zip", "x.zip")
	testPushPaths(utils.JOB, "", "/x.zip", "/artifacts/jobs/"+jobID+"/x.zip", "/x.zip")
	testPushPaths(utils.JOB, "", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/x.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/x.zip", "/long/path/to/x.zip")
	testPushPaths(utils.JOB, "y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(utils.JOB, "y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "/x.zip")
	testPushPaths(utils.JOB, "y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "/long/path/to/x.zip")
	testPushPaths(utils.JOB, "/y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(utils.JOB, "/y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "/x.zip")
	testPushPaths(utils.JOB, "/y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "/y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "/long/path/to/x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/long/path/to/x.zip")
}

func TestPullPaths(t *testing.T) {
	testPullPaths := func(category, dst, src, expDst, expSrc string) {
		resultDst, resultSrc := pullPaths(category, dst, src)
		if resultDst != expDst {
			t.Errorf("not match destination(%s) with expected(%s) for category(%s), dst(%s) and src(%s)",
				resultDst, expDst, category, dst, src)
		}
		if resultSrc != expSrc {
			t.Errorf("not match source(%s) with expected(%s) for category(%s), dst(%s) and src(%s)",
				resultSrc, expSrc, category, dst, src)
		}
	}

	jobID := "JOB_03"
	os.Setenv(utils.CategoryEnv[utils.JOB], jobID)
	testPullPaths(utils.JOB, "", "x.zip", "x.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "", "/x.zip", "x.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "", "long/path/to/x.zip", "x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "", "/long/path/to/x.zip", "x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "y.zip", "x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "y.zip", "/x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "y.zip", "long/path/to/x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "y.zip", "/long/path/to/x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "/y.zip", "x.zip", "/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/y.zip", "/x.zip", "/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/y.zip", "long/path/to/x.zip", "/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "/y.zip", "/long/path/to/x.zip", "/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "x.zip", "/long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "/x.zip", "/long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "/long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "/long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
}

func skipShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GCS tests in short mode")
	}
}

func TestGCS(t *testing.T) {
	skipShort(t)
	filename := path.Join("test", "artifact", "x.zip")
	err := writeGCS(filename, bytes.NewReader(content), time.Second*10)
	assertNilError(t, "writing to Google Cloud Storage", err)

	// TODO: test expire when it's implemented

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	err = readGCS(writer, filename)
	assertNilError(t, "reading from Google Cloud Storage", err)
	writer.Flush()
	if !bytes.Equal(b.Bytes(), content) {
		t.Errorf("downloaded content(%s) doesn't match previously uploaded(%s)", b.String(),
			string(content))
	}
	assertNilError(t, "deleting from Google Cloud Storage", err)
}

func createTmpDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "artifact")
	assertNilError(t, "creating temporary directory", err)
	if err = os.Chdir(d); err != nil {
		os.RemoveAll(d)
		t.Fatalf("failed to change to temporary directory, err: %s", err)
	}
	return d
}

func createFileWithContent(t *testing.T, name string, content []byte, expContents map[string][]byte) {
	f, err := os.Create(name)
	assertNilError(t, "creating source tmp file", err)
	defer f.Close()
	_, err = f.Write(content)
	assertNilError(t, "writing to source tmp file", err)
	expContents[path.Base(name)] = content
}

func assertFile(t *testing.T, name string) {
	ok, err := isFileGCS(name)
	assertNilError(t, "Querying a file %s at Google Cloud Storage", name, err)
	assertTrue(t, ok, "%s should be a file in Google Cloud Storage, found NOT a file", name)
}

func assertNotFile(t *testing.T, name string) {
	ok, err := isFileGCS(name)
	assertNilError(t, "Querying a file %s at Google Cloud Storage", name, err)
	assertTrue(t, !ok, "%s should NOT be a file in Google Cloud Storage, found a file", name)
}

func assertDir(t *testing.T, name string) {
	ok, err := isDirGCS(name)
	assertNilError(t, "Querying a directory %s at Google Cloud Storage", name, err)
	assertTrue(t, ok, "%s should be a directory in Google Cloud Storage, found NOT a directory", name)
}

func assertNotDir(t *testing.T, name string) {
	ok, err := isDirGCS(name)
	assertNilError(t, "Querying a directory %s at Google Cloud Storage", name, err)
	assertTrue(t, !ok, "%s should NOT be a directory in Google Cloud Storage, found a directory", name)
}

func compareFile(t *testing.T, category, dst, src string, expContent []byte, expAlready bool) {
	dst, src = pullPaths(category, dst, src)
	err := pullGCS(dst, src, false)
	if expAlready {
		assertAlreadyExists(t, "Pulling file to compare", err)
		err = pullGCS(dst, src, true)
	}
	assertNilError(t, "Pulling file to compare", err)

	c, err := ioutil.ReadFile(dst)
	assertNilError(t, "Reading destination tmp file", err)

	if !bytes.Equal(c, expContent) {
		t.Errorf("downloaded content(%s) doesn't match previously uploaded(%s)",
			string(c), string(expContent))
	}
}

func compareDir(t *testing.T, category, dst, src string, expContents map[string][]byte) {
	dst, src = pullPaths(category, dst, src)
	os.Mkdir(dst, 0777)
	err := pullGCS(dst, src, false)
	assertAlreadyExists(t, "Pulling directory to compare", err)
	err = pullGCS(dst, src, true)
	assertNilError(t, "Pulling directory to compare", err)

	// copy map, so we can remove from it
	copyContents := map[string][]byte{}
	for k, v := range expContents {
		copyContents[k] = v
	}

	filepath.Walk(dst, func(filename string, info os.FileInfo, err error) error {
		assertNilError(t, "Walking compare directory", err)
		if info.IsDir() {
			return nil
		}
		base := path.Base(filename)

		c, err := ioutil.ReadFile(filename)
		assertNilError(t, "read destination tmp file", err)
		expC, ok := copyContents[base]
		assertTrue(t, ok, "%s in an unexpected file comparing directories", filename)
		if !bytes.Equal(c, expC) {
			t.Errorf("downloaded content(%s) doesn't match previously uploaded(%s)",
				string(c), string(expC))
		}
		delete(copyContents, base)
		return nil
	})

	if len(copyContents) > 0 {
		t.Errorf("some expected items hasn't been downloaded, orig len: %d; stack: %s",
			len(expContents), debug.Stack())
		for k, v := range copyContents {
			t.Errorf("item name: %s, value: %s", k, string(v))
		}
	}
}

func TestGCSOverwrite(t *testing.T) {
	skipShort(t)
	delDirGCS("/artifacts")

	compareD := createTmpDir(t)
	// defer os.RemoveAll(compareD)
	compareDDest := path.Join(compareD, "check")

	d := createTmpDir(t)
	defer os.RemoveAll(d)

	d2 := path.Join(d, "dir")
	os.Mkdir(d2, 0777)

	expContents := map[string][]byte{}
	srcFilename := path.Join(d2, "x.txt")
	createFileWithContent(t, srcFilename, content, expContents)

	category := utils.JOB
	gcsFilenameX := "x"
	dst, src := pushPaths(category, gcsFilenameX, srcFilename)
	assertNotFile(t, dst)
	assertNotDir(t, dst)

	err := pushGCS(dst, src, "100", false)
	assertNilError(t, "push file to Google Cloud Storage", err)
	assertFile(t, dst)
	assertNotDir(t, dst)
	compareFile(t, category, compareDDest, gcsFilenameX, content, false)

	srcFilename2 := path.Join(d2, "x")
	createFileWithContent(t, srcFilename2, content2, expContents)

	// trying to overwrite file with a file without force; expectation: fails
	dst, src = pushPaths(category, "", srcFilename2)
	err = pushGCS(dst, src, "100", false)
	assertAlreadyExists(t, "overwriting file with file without force", err)
	assertFile(t, dst)
	assertNotDir(t, dst)
	compareFile(t, category, compareDDest, gcsFilenameX, content, true)

	// trying to overwrite file with a file with force; expectation: succeeds
	dst, src = pushPaths(category, "", srcFilename2)
	err = pushGCS(dst, src, "100", true)
	assertNilError(t, "force push file to Google Cloud Storage", err)
	assertFile(t, dst)
	assertNotDir(t, dst)
	compareFile(t, category, compareDDest, gcsFilenameX, content2, true)

	// trying to overwrite file with a directory without force; expectation: fails
	dst, src = pushPaths(category, gcsFilenameX, d2)
	err = pushGCS(dst, src, "100", false)
	assertAlreadyExists(t, "overwriting file with directory without force", err)
	assertFile(t, dst)
	assertNotDir(t, dst)
	compareFile(t, category, compareDDest, gcsFilenameX, content2, true)

	// trying to overwrite file with a directory with force; expectation: succeeds
	dst, src = pushPaths(category, gcsFilenameX, d2)
	err = pushGCS(dst, src, "100", true)
	assertNilError(t, "overwriting file with directory with force", err)
	assertNotFile(t, dst)
	assertDir(t, dst)
	compareDir(t, category, compareDDest, gcsFilenameX, expContents)

	// new directory content
	expContents2 := map[string][]byte{}
	d2 = path.Join(d, "dir2")
	os.Mkdir(d2, 0777)
	srcFilename = path.Join(d2, "y.txt")
	createFileWithContent(t, srcFilename, content2, expContents2)
	srcFilename = path.Join(d2, "z.txt")
	createFileWithContent(t, srcFilename, content, expContents2)

	// trying to overwrite directory with a directory without force; expectation: fail
	dst, src = pushPaths(category, gcsFilenameX, d2)
	err = pushGCS(dst, src, "100", false)
	assertAlreadyExists(t, "overwriting directory with directory without force", err)
	assertNotFile(t, dst)
	assertDir(t, dst)
	compareDir(t, category, compareDDest, gcsFilenameX, expContents)

	// trying to overwrite directory with a directory with force; expectation: success
	dst, src = pushPaths(category, gcsFilenameX, d2)
	err = pushGCS(dst, src, "100", true)
	assertNilError(t, "overwriting directory with directory with force", err)
	assertNotFile(t, dst)
	assertDir(t, dst)
	compareDir(t, category, compareDDest, gcsFilenameX, expContents2)

	// trying to overwrite directory with a file without force; expectation: fails
	dst, src = pushPaths(category, "", srcFilename2)
	err = pushGCS(dst, src, "100", false)
	assertAlreadyExists(t, "overwriting directory with file without force", err)
	assertNotFile(t, dst)
	assertDir(t, dst)
	compareDir(t, category, compareDDest, gcsFilenameX, expContents2)

	// trying to overwrite directory with a file with force; expectation: succeeds
	dst, src = pushPaths(category, gcsFilenameX, srcFilename)
	err = pushGCS(dst, src, "100", true)
	assertNilError(t, "force push file to Google Cloud Storage", err)
	assertFile(t, dst)
	assertNotDir(t, dst)
	compareFile(t, category, compareDDest, gcsFilenameX, content, true)

	filename := yankPath(category, gcsFilenameX)
	err = yankGCS(filename)
	assertNilError(t, "yank file from Google Cloud Storage", err)
	assertNotFile(t, filename)
	assertNotDir(t, filename)
}
