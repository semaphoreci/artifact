package pathutil

import (
	"os"
	"testing"
)

func TestMissingCategoryID(t *testing.T) {
	check := func(category, categoryID string, expOk bool) {
		err := InitPathID(category, categoryID)
		if expOk != (err == nil) {
			t.Errorf("not match result(%s), expected(%t) for missing categoryID(%s), cat: %s",
				err, expOk, categoryID, category)
		}
	}

	p := "some_project"
	check(PROJECT, "", false)
	check(PROJECT, p, true)
	os.Setenv(CategoryEnv[PROJECT], p)
	check(PROJECT, "", true)
	check(PROJECT, p, true)
	os.Setenv(CategoryEnv[PROJECT], "")
	check(PROJECT, "", false)

	w := "some_workflow"
	check(WORKFLOW, "", false)
	check(WORKFLOW, w, true)
	os.Setenv(CategoryEnv[WORKFLOW], w)
	check(WORKFLOW, "", true)
	check(WORKFLOW, w, true)
	os.Setenv(CategoryEnv[WORKFLOW], "")
	check(WORKFLOW, "", false)

	j := "some_job"
	check(JOB, "", false)
	check(JOB, j, true)
	os.Setenv(CategoryEnv[JOB], j)
	check(JOB, "", true)
	check(JOB, j, true)
	os.Setenv(CategoryEnv[JOB], "")
	check(JOB, "", false)
}

func TestPrefixedPathEmptyDefault(t *testing.T) {
	testPrefixedPath := func(category, filepath, expected string) {
		InitPathID(category, "")
		result := PrefixedPath(filepath)
		if result != expected {
			t.Errorf("not match result(%s) with expected(%s) for category(%s) and filepath(%s)",
				result, expected, category, filepath)
		}
	}

	projectID := "PR_01"
	os.Setenv(CategoryEnv[PROJECT], projectID)
	workflowID := "WF_02"
	os.Setenv(CategoryEnv[WORKFLOW], workflowID)
	jobID := "JOB_03"
	os.Setenv(CategoryEnv[JOB], jobID)
	testPrefixedPath(PROJECT, ".", "artifacts/projects/"+projectID)
	testPrefixedPath(PROJECT, "x.zip", "artifacts/projects/"+projectID+"/x.zip")
	testPrefixedPath(PROJECT, "y.zip", "artifacts/projects/"+projectID+"/y.zip")
	testPrefixedPath(PROJECT, "tmp/x.zip", "artifacts/projects/"+projectID+"/tmp/x.zip")
	testPrefixedPath(PROJECT, "/tmp/x.zip", "artifacts/projects/"+projectID+"/tmp/x.zip")
	testPrefixedPath(WORKFLOW, "x.zip", "artifacts/workflows/"+workflowID+"/x.zip")
	testPrefixedPath(WORKFLOW, "path/to/the/deep/x.zip", "artifacts/workflows/"+workflowID+
		"/path/to/the/deep/x.zip")
	testPrefixedPath(JOB, "x.zip", "artifacts/jobs/"+jobID+"/x.zip")
}

func TestPrefixedPathSetDefault(t *testing.T) {
	testPrefixedPath := func(category, filepath, expected string) {
		InitPathID(category, "fixed")
		result := PrefixedPath(filepath)
		if result != expected {
			t.Errorf("not match result(%s) with expected(%s) for category(%s) and filepath(%s)",
				result, expected, category, filepath)
		}
	}

	projectID := "PR_01"
	os.Setenv(CategoryEnv[PROJECT], projectID)
	workflowID := "WF_02"
	os.Setenv(CategoryEnv[WORKFLOW], workflowID)
	jobID := "JOB_03"
	os.Setenv(CategoryEnv[JOB], jobID)
	fixed := "fixed"
	testPrefixedPath(JOB, ".", "artifacts/jobs/"+fixed)
	testPrefixedPath(JOB, "x.zip", "artifacts/jobs/"+fixed+"/x.zip")
	testPrefixedPath(JOB, "y.zip", "artifacts/jobs/"+fixed+"/y.zip")
	testPrefixedPath(JOB, "tmp/x.zip", "artifacts/jobs/"+fixed+"/tmp/x.zip")
	testPrefixedPath(JOB, "/tmp/x.zip", "artifacts/jobs/"+fixed+"/tmp/x.zip")
	testPrefixedPath(PROJECT, "x.zip", "artifacts/projects/"+fixed+"/x.zip")
	testPrefixedPath(PROJECT, "path/to/the/deep/x.zip", "artifacts/projects/"+fixed+
		"/path/to/the/deep/x.zip")
	testPrefixedPath(WORKFLOW, "x.zip", "artifacts/workflows/"+fixed+"/x.zip")
}

func TestPathFromSource(t *testing.T) {
	testPathFromSource := func(dst, src, expDst string) {
		result := PathFromSource(dst, src)
		if result != expDst {
			t.Errorf("not match result(%s) with expected(%s) for dst(%s) and src(%s)",
				result, expDst, dst, src)
		}
	}

	testPathFromSource("", "/long/path/to/source", "source")
	testPathFromSource("", "/long/path/to/.source", ".source")
	testPathFromSource("", "long/path/to/source", "source")
	testPathFromSource("", "long/path/to/.source", ".source")
	testPathFromSource("destination", "/long/path/to/source", "destination")
	testPathFromSource(".destination", "/long/path/to/source", ".destination")
	testPathFromSource("destination", "/long/path/to/.source", "destination")
	testPathFromSource(".destination", "/long/path/to/.source", ".destination")
	testPathFromSource("destination", "long/path/to/source", "destination")
	testPathFromSource(".destination", "long/path/to/source", ".destination")
	testPathFromSource("destination", "long/path/to/.source", "destination")
	testPathFromSource(".destination", "long/path/to/.source", ".destination")
	testPathFromSource("long/path/to/destination", "long/path/to/source",
		"long/path/to/destination")
	testPathFromSource(".long/path/to/destination", "long/path/to/source",
		".long/path/to/destination")
	testPathFromSource("long/path/to/destination", ".long/path/to/source",
		"long/path/to/destination")
	testPathFromSource(".long/path/to/destination", ".long/path/to/source",
		".long/path/to/destination")
	testPathFromSource("/long/path/to/destination", "long/path/to/source",
		"/long/path/to/destination")
	testPathFromSource("/.long/path/to/destination", "long/path/to/source",
		"/.long/path/to/destination")
	testPathFromSource("/long/path/to/destination", ".long/path/to/source",
		"/long/path/to/destination")
	testPathFromSource("/.long/path/to/destination", ".long/path/to/source",
		"/.long/path/to/destination")
	testPathFromSource("./long/path/to/destination", "long/path/to/source",
		"./long/path/to/destination")
	testPathFromSource("./.long/path/to/destination", "long/path/to/source",
		"./.long/path/to/destination")
	testPathFromSource("./long/path/to/destination", ".long/path/to/source",
		"./long/path/to/destination")
	testPathFromSource("./.long/path/to/destination", ".long/path/to/source",
		"./.long/path/to/destination")
}

func TestToRelative(t *testing.T) {
	testToRelative := func(src, expected string) {
		result := ToRelative(src)
		if result != expected {
			t.Errorf("not match result(%s) with expected(%s) for src(%s)",
				result, expected, src)
		}
	}

	testToRelative("", "")
	testToRelative("./../source", "source")
	testToRelative("./../.source", ".source")
	testToRelative("./../source/..", "")
	testToRelative("./../source/../longer", "longer")
	testToRelative("./../source/../longer/", "longer")
	testToRelative("./../source/../.longer/", ".longer")
	testToRelative("./../source/../longer/.", "longer")
	testToRelative("./../source/../.longer/.", ".longer")
	testToRelative("./../.source/../longer/.", "longer")
	testToRelative("./../.source/../.longer/.", ".longer")
	testToRelative("source", "source")
	testToRelative(".source", ".source")
	testToRelative("/source", "source")
	testToRelative("./source", "source")
	testToRelative("/.source", ".source")
	testToRelative("long/path/to/source", "long/path/to/source")
	testToRelative(".long/path/to/source", ".long/path/to/source")
	testToRelative("/long/path/to/source", "long/path/to/source")
	testToRelative("/.long/path/to/source", ".long/path/to/source")
	testToRelative("./.long/path/to/source", ".long/path/to/source")
}
