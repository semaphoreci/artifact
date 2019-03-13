package cmd

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"

	"path"
	"testing"
	"time"

	"github.com/semaphoreci/artifact/cmd/utils"
)

var (
	content = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
)

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
	testPushPaths(utils.JOB, "", "/x.zip", "/artifacts/jobs/"+jobID+"/x.zip", "x.zip")
	testPushPaths(utils.JOB, "", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/x.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/x.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(utils.JOB, "y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(utils.JOB, "y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "/y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(utils.JOB, "/y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(utils.JOB, "/y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "/y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(utils.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "/artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
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
	testPullPaths(utils.JOB, "/y.zip", "x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/y.zip", "/x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/y.zip", "long/path/to/x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "/y.zip", "/long/path/to/x.zip", "y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(utils.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "/artifacts/jobs/"+jobID+"/long/path/to/x.zip")
}

func TestGCS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GCS tests in short mode")
	}
	filename := path.Join("test", "artifact", "x.zip")
	err := writeGCS(filename, bytes.NewReader(content), time.Second*10)
	if err != nil {
		t.Fatalf("failed to write to Google Cloud Storage, err: %s", err)
	}

	// TODO: test expire when it's implemented

	if err = writeGCS(filename, bytes.NewReader(content), time.Second*10); err == nil {
		t.Fatalf("would be able to overwrite object in Google Cloud Storage")
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err = readGCS(writer, filename); err != nil {
		t.Fatalf("failed to read from Google Cloud Storage, err: %s", err)
	}
	writer.Flush()
	if !bytes.Equal(b.Bytes(), content) {
		t.Errorf("downloaded content(%s) doesn't match previously uploaded(%s)", b.String(), string(content))
	}
	if err = delGCS(filename); err != nil {
		t.Fatalf("failed to delete from to Google Cloud Storage, err: %s", err)
	}
}

func TestGCSAdvanced(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GCS tests in short mode")
	}

	d, err := ioutil.TempDir("", "artifact")
	if err != nil {
		t.Fatalf("failed to create temporary directory, err: %s", err)
	}
	defer os.RemoveAll(d)
	if err = os.Chdir(d); err != nil {
		t.Fatalf("failed to change to temporary directory, err: %s", err)
	}

	srcFilename := "x.txt"
	f, err := os.Create(srcFilename)
	if err != nil {
		t.Fatalf("failed to create source tmp file, err: %s", err)
	}
	if _, err = f.Write(content); err != nil {
		f.Close()
		t.Fatalf("failed to write to source tmp file, err: %s", err)
	}
	f.Close()

	gcsFilename := "debug/y.txt"
	if _, _, err = pushFileGCS(utils.JOB, "/"+gcsFilename, srcFilename, "10"); err != nil {
		t.Fatalf("failed to push to Google Cloud Storage, err: %s", err)
	}

	// TODO: test expire when it's implemented

	dstFilename := "test/z.txt"
	if _, _, err = pullFileGCS(utils.JOB, "/"+dstFilename, gcsFilename); err != nil {
		t.Fatalf("failed to pull from Google Cloud Storage, err: %s", err)
	}

	c2, err := ioutil.ReadFile(dstFilename)
	if err != nil {
		t.Fatalf("failed to read destination tmp file, err: %s", err)
	}

	if !bytes.Equal(c2, content) {
		t.Errorf("downloaded content(%s) doesn't match previously uploaded(%s)", string(c2), string(content))
	}
	if _, err = yankFileGCS(utils.JOB, "/"+gcsFilename); err != nil {
		t.Fatalf("failed to yank from to Google Cloud Storage, err: %s", err)
	}
}
