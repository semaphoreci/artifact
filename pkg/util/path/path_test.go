package pathutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingCategoryID(t *testing.T) {
	check := func(category, categoryID string, expOk bool) {
		err := InitPathID(category, categoryID)
		assert.Equal(t, expOk, err == nil, categoryID, category)
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
	check := func(category, filepath, expected string) {
		InitPathID(category, "")
		result := PrefixedPath(filepath)
		assert.Equal(t, expected, result, category, filepath)
	}

	projectID := "PR_01"
	os.Setenv(CategoryEnv[PROJECT], projectID)
	workflowID := "WF_02"
	os.Setenv(CategoryEnv[WORKFLOW], workflowID)
	jobID := "JOB_03"
	os.Setenv(CategoryEnv[JOB], jobID)
	check(PROJECT, ".", "artifacts/projects/"+projectID)
	check(PROJECT, "x.zip", "artifacts/projects/"+projectID+"/x.zip")
	check(PROJECT, "y.zip", "artifacts/projects/"+projectID+"/y.zip")
	check(PROJECT, "tmp/x.zip", "artifacts/projects/"+projectID+"/tmp/x.zip")
	check(PROJECT, "/tmp/x.zip", "artifacts/projects/"+projectID+"/tmp/x.zip")
	check(WORKFLOW, "x.zip", "artifacts/workflows/"+workflowID+"/x.zip")
	check(WORKFLOW, "path/to/the/deep/x.zip", "artifacts/workflows/"+workflowID+
		"/path/to/the/deep/x.zip")
	check(JOB, "x.zip", "artifacts/jobs/"+jobID+"/x.zip")
}

func TestPrefixedPathSetDefault(t *testing.T) {
	check := func(category, filepath, expected string) {
		InitPathID(category, "fixed")
		result := PrefixedPath(filepath)
		assert.Equal(t, expected, result, category, filepath)
	}

	projectID := "PR_01"
	os.Setenv(CategoryEnv[PROJECT], projectID)
	workflowID := "WF_02"
	os.Setenv(CategoryEnv[WORKFLOW], workflowID)
	jobID := "JOB_03"
	os.Setenv(CategoryEnv[JOB], jobID)
	fixed := "fixed"
	check(JOB, ".", "artifacts/jobs/"+fixed)
	check(JOB, "x.zip", "artifacts/jobs/"+fixed+"/x.zip")
	check(JOB, "y.zip", "artifacts/jobs/"+fixed+"/y.zip")
	check(JOB, "tmp/x.zip", "artifacts/jobs/"+fixed+"/tmp/x.zip")
	check(JOB, "/tmp/x.zip", "artifacts/jobs/"+fixed+"/tmp/x.zip")
	check(PROJECT, "x.zip", "artifacts/projects/"+fixed+"/x.zip")
	check(PROJECT, "path/to/the/deep/x.zip", "artifacts/projects/"+fixed+"/path/to/the/deep/x.zip")
	check(WORKFLOW, "x.zip", "artifacts/workflows/"+fixed+"/x.zip")
}

func TestPathFromSource(t *testing.T) {
	check := func(dst, src, expDst string) {
		result := PathFromSource(dst, src)
		assert.Equal(t, expDst, result, dst, src)
	}

	check("", "/long/path/to/source", "source")
	check("", "/long/path/to/.source", ".source")
	check("", "long/path/to/source", "source")
	check("", "long/path/to/.source", ".source")
	check("destination", "/long/path/to/source", "destination")
	check(".destination", "/long/path/to/source", ".destination")
	check("destination", "/long/path/to/.source", "destination")
	check(".destination", "/long/path/to/.source", ".destination")
	check("destination", "long/path/to/source", "destination")
	check(".destination", "long/path/to/source", ".destination")
	check("destination", "long/path/to/.source", "destination")
	check(".destination", "long/path/to/.source", ".destination")
	check("long/path/to/destination", "long/path/to/source", "long/path/to/destination")
	check(".long/path/to/destination", "long/path/to/source", ".long/path/to/destination")
	check("long/path/to/destination", ".long/path/to/source", "long/path/to/destination")
	check(".long/path/to/destination", ".long/path/to/source", ".long/path/to/destination")
	check("/long/path/to/destination", "long/path/to/source", "/long/path/to/destination")
	check("/.long/path/to/destination", "long/path/to/source", "/.long/path/to/destination")
	check("/long/path/to/destination", ".long/path/to/source", "/long/path/to/destination")
	check("/.long/path/to/destination", ".long/path/to/source", "/.long/path/to/destination")
	check("./long/path/to/destination", "long/path/to/source", "./long/path/to/destination")
	check("./.long/path/to/destination", "long/path/to/source", "./.long/path/to/destination")
	check("./long/path/to/destination", ".long/path/to/source", "./long/path/to/destination")
	check("./.long/path/to/destination", ".long/path/to/source", "./.long/path/to/destination")
}

func TestToRelative(t *testing.T) {
	check := func(src, expected string) {
		result := ToRelative(src)
		assert.Equal(t, expected, result, src)
	}

	check("", "")
	check("./../source", "source")
	check("./../.source", ".source")
	check("./../source/..", "")
	check("./../source/../longer", "longer")
	check("./../source/../longer/", "longer")
	check("./../source/../.longer/", ".longer")
	check("./../source/../longer/.", "longer")
	check("./../source/../.longer/.", ".longer")
	check("./../.source/../longer/.", "longer")
	check("./../.source/../.longer/.", ".longer")
	check("source", "source")
	check(".source", ".source")
	check("/source", "source")
	check("./source", "source")
	check("/.source", ".source")
	check("long/path/to/source", "long/path/to/source")
	check(".long/path/to/source", ".long/path/to/source")
	check("/long/path/to/source", "long/path/to/source")
	check("/.long/path/to/source", ".long/path/to/source")
	check("./.long/path/to/source", ".long/path/to/source")
}
