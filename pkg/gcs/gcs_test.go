package gcs

import (
	"os"
	"testing"

	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
)

var (
	content  = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	content2 = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam lacus massa, porttitor non euismod vel, volutpat eget metus. Maecenas finibus interdum ante id rhoncus. Mauris sodales congue volutpat. Integer scelerisque elit nec lectus varius luctus. Ut tempor orci at tellus facilisis interdum. In scelerisque nec sem vitae euismod. Suspendisse nibh nulla, egestas varius tortor quis, hendrerit cursus urna. Maecenas eu risus ligula. Sed eu tortor orci. Donec mattis cursus gravida.")
)

func TestPushPathsEmptyDefault(t *testing.T) {
	testPushPaths := func(category, dst, src, expDst, expSrc string) {
		pathutil.InitPathID(category, "")
		resultDst, resultSrc := PushPaths(dst, src)
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
	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	testPushPaths(pathutil.JOB, "", "x.zip", "artifacts/jobs/"+jobID+"/x.zip", "x.zip")
	testPushPaths(pathutil.JOB, "", "/x.zip", "artifacts/jobs/"+jobID+"/x.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/x.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/x.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/long/path/to/x.zip")
}

func TestPushPathsSetDefault(t *testing.T) {
	testPushPaths := func(category, dst, src, expDst, expSrc string) {
		pathutil.InitPathID(category, "fixed")
		resultDst, resultSrc := PushPaths(dst, src)
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
	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	fixed := "fixed"
	testPushPaths(pathutil.JOB, "", "x.zip", "artifacts/jobs/"+fixed+"/x.zip", "x.zip")
	testPushPaths(pathutil.JOB, "", "/x.zip", "artifacts/jobs/"+fixed+"/x.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/x.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/x.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths(pathutil.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/long/path/to/x.zip")
}

func TestPullPathsEmptyDefault(t *testing.T) {
	testPullPaths := func(category, dst, src, expDst, expSrc string) {
		pathutil.InitPathID(category, "")
		resultDst, resultSrc := PullPaths(dst, src)
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
	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	testPullPaths(pathutil.JOB, "", "x.zip", "x.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "", "/x.zip", "x.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "", "long/path/to/x.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "", "/long/path/to/x.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "/long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "/x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "/long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
}

func TestPullPathsSetDefault(t *testing.T) {
	testPullPaths := func(category, dst, src, expDst, expSrc string) {
		pathutil.InitPathID(category, "fixed")
		resultDst, resultSrc := PullPaths(dst, src)
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
	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	fixed := "fixed"
	testPullPaths(pathutil.JOB, "", "x.zip", "x.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "", "/x.zip", "x.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "", "long/path/to/x.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "", "/long/path/to/x.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "y.zip", "/long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "/x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/y.zip", "/long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths(pathutil.JOB, "/long/path/to/y.zip", "/long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
}
