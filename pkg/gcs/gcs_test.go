package gcs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	httpmock "github.com/jarcoal/httpmock"
	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
)

const (
	fixed    = "fixed"
	jobID    = "JOB_03"
	category = pathutil.JOB
)

var (
	content  = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	content2 = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam lacus massa, porttitor non euismod vel, volutpat eget metus. Maecenas finibus interdum ante id rhoncus. Mauris sodales congue volutpat. Integer scelerisque elit nec lectus varius luctus. Ut tempor orci at tellus facilisis interdum. In scelerisque nec sem vitae euismod. Suspendisse nibh nulla, egestas varius tortor quis, hendrerit cursus urna. Maecenas eu risus ligula. Sed eu tortor orci. Donec mattis cursus gravida.")

	reqURL = os.Getenv("SEMAPHORE_ORGANIZATION_URL") + gatewayAPIBase
)

func TestRetryHTTPReqSuccess(t *testing.T) {
	httpmock.Activate()
	numOfTries := 0
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", reqURL,
		func(req *http.Request) (*http.Response, error) {
			if numOfTries < 3 {
				numOfTries++
				return httpmock.NewStringResponse(500, ""), nil
			}
			jsonResp := make(map[string]interface{})
			jsonData := []byte(fmt.Sprintf(`{"Urls":[{"URL": %v,"Method":"PUT"]}`, reqURL))
			json.Unmarshal(jsonData, &jsonResp)
			resp, _ := httpmock.NewJsonResponse(200, jsonResp)
			return resp, nil
		},
	)

	request := &GenerateSignedURLsRequest{Paths: []string{"/test/path"}}
	request.Type = generateSignedURLsRequestPUSH
	var x GenerateSignedURLsResponse
	if fail := errutil.RetryOnFailure("", func() bool {
		return handleHTTPReq(request, &x)
	}); fail == true {
		t.Errorf("Failed to preform request")
	}
}

func TestRetryableHTTPReqFailure(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", reqURL,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(500, ""), nil
		},
	)

	request := &GenerateSignedURLsRequest{Paths: []string{"/test/path"}}
	request.Type = generateSignedURLsRequestPUSH
	var x GenerateSignedURLsResponse
	if fail := errutil.RetryOnFailure("", func() bool {
		return handleHTTPReq(request, &x)
	}); fail == false {
		t.Errorf("Result must be Failure")
	}
}

func TestPushPathsEmptyDefault(t *testing.T) {
	testPushPaths := func(dst, src, expDst, expSrc string) {
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

	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	testPushPaths("", "x.zip", "artifacts/jobs/"+jobID+"/x.zip", "x.zip")
	testPushPaths("", "/x.zip", "artifacts/jobs/"+jobID+"/x.zip", "/x.zip")
	testPushPaths("", "./x.zip", "artifacts/jobs/"+jobID+"/x.zip", "x.zip")
	testPushPaths("", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/x.zip", "long/path/to/x.zip")
	testPushPaths("", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/x.zip", "/long/path/to/x.zip")
	testPushPaths("", "./long/path/to/x.zip", "artifacts/jobs/"+jobID+"/x.zip", "long/path/to/x.zip")
	testPushPaths("y.zip", "x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths("y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/x.zip")
	testPushPaths("y.zip", "./x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths("y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths("y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/long/path/to/x.zip")
	testPushPaths("y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths("/y.zip", "x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths("/y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/x.zip")
	testPushPaths("/y.zip", "./x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths("/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths("/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/long/path/to/x.zip")
	testPushPaths("/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths("./y.zip", "x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths("./y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/x.zip")
	testPushPaths("./y.zip", "./x.zip", "artifacts/jobs/"+jobID+"/y.zip", "x.zip")
	testPushPaths("./y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths("./y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "/long/path/to/x.zip")
	testPushPaths("./y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+jobID+"/y.zip", "long/path/to/x.zip")
	testPushPaths("long/path/to/y.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths("long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/x.zip")
	testPushPaths("long/path/to/y.zip", "./x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths("long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths("long/path/to/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("/long/path/to/y.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths("/long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/x.zip")
	testPushPaths("/long/path/to/y.zip", "./x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths("/long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("/long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths("/long/path/to/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("./long/path/to/y.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths("./long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/x.zip")
	testPushPaths("./long/path/to/y.zip", "./x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "x.zip")
	testPushPaths("./long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("./long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths("./long/path/to/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+jobID+"/long/path/to/y.zip", "long/path/to/x.zip")
}

func TestPushPathsSetDefault(t *testing.T) {
	testPushPaths := func(dst, src, expDst, expSrc string) {
		pathutil.InitPathID(category, fixed)
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

	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	testPushPaths("", "x.zip", "artifacts/jobs/"+fixed+"/x.zip", "x.zip")
	testPushPaths("", "/x.zip", "artifacts/jobs/"+fixed+"/x.zip", "/x.zip")
	testPushPaths("", "./x.zip", "artifacts/jobs/"+fixed+"/x.zip", "x.zip")
	testPushPaths("", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/x.zip", "long/path/to/x.zip")
	testPushPaths("", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/x.zip", "/long/path/to/x.zip")
	testPushPaths("", "./long/path/to/x.zip", "artifacts/jobs/"+fixed+"/x.zip", "long/path/to/x.zip")
	testPushPaths("y.zip", "x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths("y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/x.zip")
	testPushPaths("y.zip", "./x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths("y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths("y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/long/path/to/x.zip")
	testPushPaths("y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths("/y.zip", "x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths("/y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/x.zip")
	testPushPaths("/y.zip", "./x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths("/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths("/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/long/path/to/x.zip")
	testPushPaths("/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths("./y.zip", "x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths("./y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/x.zip")
	testPushPaths("./y.zip", "./x.zip", "artifacts/jobs/"+fixed+"/y.zip", "x.zip")
	testPushPaths("./y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths("./y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "/long/path/to/x.zip")
	testPushPaths("./y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+fixed+"/y.zip", "long/path/to/x.zip")
	testPushPaths("long/path/to/y.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths("long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/x.zip")
	testPushPaths("long/path/to/y.zip", "./x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths("long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths("long/path/to/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("/long/path/to/y.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths("/long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/x.zip")
	testPushPaths("/long/path/to/y.zip", "./x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths("/long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("/long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths("/long/path/to/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("./long/path/to/y.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths("./long/path/to/y.zip", "/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/x.zip")
	testPushPaths("./long/path/to/y.zip", "./x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "x.zip")
	testPushPaths("./long/path/to/y.zip", "long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
	testPushPaths("./long/path/to/y.zip", "/long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "/long/path/to/x.zip")
	testPushPaths("./long/path/to/y.zip", "./long/path/to/x.zip", "artifacts/jobs/"+fixed+"/long/path/to/y.zip", "long/path/to/x.zip")
}

func TestPullPathsEmptyDefault(t *testing.T) {
	testPullPaths := func(dst, src, expDst, expSrc string) {
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

	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	testPullPaths("", "x.zip", "x.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("", "/x.zip", "x.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("", "./x.zip", "x.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("", "long/path/to/x.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("", "/long/path/to/x.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("", "./long/path/to/x.zip", "x.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("y.zip", "x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("y.zip", "/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("y.zip", "./x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("y.zip", "long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("y.zip", "/long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("y.zip", "./long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("/y.zip", "x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("/y.zip", "/x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("/y.zip", "./x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("/y.zip", "long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("/y.zip", "/long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("/y.zip", "./long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("./y.zip", "x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("./y.zip", "/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("./y.zip", "./x.zip", "y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("./y.zip", "long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("./y.zip", "/long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("./y.zip", "./long/path/to/x.zip", "y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("long/path/to/y.zip", "./x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("long/path/to/y.zip", "./long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("/long/path/to/y.zip", "x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("/long/path/to/y.zip", "/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("/long/path/to/y.zip", "./x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("/long/path/to/y.zip", "long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("/long/path/to/y.zip", "/long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("/long/path/to/y.zip", "./long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("./long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("./long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("./long/path/to/y.zip", "./x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/x.zip")
	testPullPaths("./long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("./long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
	testPullPaths("./long/path/to/y.zip", "./long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+jobID+"/long/path/to/x.zip")
}

func TestPullPathsSetDefault(t *testing.T) {
	testPullPaths := func(dst, src, expDst, expSrc string) {
		pathutil.InitPathID(category, fixed)
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

	os.Setenv(pathutil.CategoryEnv[pathutil.JOB], jobID)
	testPullPaths("", "x.zip", "x.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("", "/x.zip", "x.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("", "./x.zip", "x.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("", "long/path/to/x.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("", "/long/path/to/x.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("", "./long/path/to/x.zip", "x.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("y.zip", "x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("y.zip", "/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("y.zip", "./x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("y.zip", "long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("y.zip", "/long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("y.zip", "./long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("/y.zip", "x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("/y.zip", "/x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("/y.zip", "./x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("/y.zip", "long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("/y.zip", "/long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("/y.zip", "./long/path/to/x.zip", "/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("./y.zip", "x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("./y.zip", "/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("./y.zip", "./x.zip", "y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("./y.zip", "long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("./y.zip", "/long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("./y.zip", "./long/path/to/x.zip", "y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("long/path/to/y.zip", "./x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("long/path/to/y.zip", "./long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("/long/path/to/y.zip", "x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("/long/path/to/y.zip", "/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("/long/path/to/y.zip", "./x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("/long/path/to/y.zip", "long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("/long/path/to/y.zip", "/long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("/long/path/to/y.zip", "./long/path/to/x.zip", "/long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("./long/path/to/y.zip", "x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("./long/path/to/y.zip", "/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("./long/path/to/y.zip", "./x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPullPaths("./long/path/to/y.zip", "long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("./long/path/to/y.zip", "/long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
	testPullPaths("./long/path/to/y.zip", "./long/path/to/x.zip", "long/path/to/y.zip", "artifacts/jobs/"+fixed+"/long/path/to/x.zip")
}
