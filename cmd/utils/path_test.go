package utils

import (
	"os"
	"testing"
)

func TestPrefixedPath(t *testing.T) {
	testPrefixedPath := func(category, filepath, expected string) {
		result := PrefixedPath(category, filepath)
		if result != expected {
			t.Errorf("not match result(%s) with expected(%s) for category(%s) and filepath(%s)",
				result, expected, category, filepath)
		}
	}

	projectID := "PR_01"
	os.Setenv(categoryEnv[PROJECT], projectID)
	workflowID := "WF_02"
	os.Setenv(categoryEnv[WORKFLOW], workflowID)
	jobID := "JOB_03"
	os.Setenv(categoryEnv[JOB], jobID)
	testPrefixedPath(PROJECT, "x.zip", "/artifacts/projects/"+projectID+"/x.zip")
	testPrefixedPath(PROJECT, "y.zip", "/artifacts/projects/"+projectID+"/y.zip")
	testPrefixedPath(PROJECT, "tmp/x.zip", "/artifacts/projects/"+projectID+"/tmp/x.zip")
	testPrefixedPath(WORKFLOW, "x.zip", "/artifacts/workflows/"+workflowID+"/x.zip")
	testPrefixedPath(WORKFLOW, "path/to/the/deep/x.zip", "/artifacts/workflows/"+workflowID+"/path/to/the/deep/x.zip")
	testPrefixedPath(JOB, "x.zip", "/artifacts/jobs/"+jobID+"/x.zip")
}

func TestPathFromSource(t *testing.T) {
	testPathFromSource := func(dst, src, expected string) {
		result := PathFromSource(dst, src)
		if result != expected {
			t.Errorf("not match result(%s) with expected(%s) for dst(%s) and src(%s)",
				result, expected, dst, src)
		}
	}

	testPathFromSource("", "/path/to/test", "test")
	testPathFromSource("", "path/to/test", "test")
	testPathFromSource("foo", "/path/to/test", "foo")
	testPathFromSource("foo", "path/to/test", "foo")
}
