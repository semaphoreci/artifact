package files

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	ResourceType        string
	EnvironmentVariable string
}

type resolvedPathAssertion struct {
	InSrc  string
	InDst  string
	OutSrc string
	OutDst string
}

func Test__Resolve(t *testing.T) {
	testCases := []testCase{
		{
			ResourceType:        ResourceTypeProject,
			EnvironmentVariable: "SEMAPHORE_PROJECT_ID",
		},
		{
			ResourceType:        ResourceTypeWorkflow,
			EnvironmentVariable: "SEMAPHORE_WORKFLOW_ID",
		},
		{
			ResourceType:        ResourceTypeJob,
			EnvironmentVariable: "SEMAPHORE_JOB_ID",
		},
	}

	for _, testCase := range testCases {
		runForResourceType(t, testCase)
	}
}

func runForResourceType(t *testing.T, testCase testCase) {
	t.Run(testCase.ResourceType+" uses environment variable by default", func(t *testing.T) {
		os.Setenv(testCase.EnvironmentVariable, "1")

		resolver, err := NewPathResolver(testCase.ResourceType, "")
		assert.Nil(t, err)
		assert.Equal(t, resolver.ResourceIdentifier, "1")
	})

	t.Run(testCase.ResourceType+" uses override", func(t *testing.T) {
		os.Setenv(testCase.EnvironmentVariable, "1")

		resolver, err := NewPathResolver(testCase.ResourceType, "2")
		assert.Nil(t, err)
		assert.Equal(t, resolver.ResourceIdentifier, "2")
	})

	t.Run(testCase.ResourceType+" missing resource id", func(t *testing.T) {
		os.Setenv(testCase.EnvironmentVariable, "")

		_, err := NewPathResolver(testCase.ResourceType, "")
		assert.NotNil(t, err)
	})

	t.Run(testCase.ResourceType+" resolve paths for push", func(t *testing.T) {
		os.Setenv(testCase.EnvironmentVariable, "1")
		resolver, _ := NewPathResolver(testCase.ResourceType, "")

		assertions := []resolvedPathAssertion{
			{InDst: "", InSrc: "x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "", InSrc: "/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutSrc: "/x.zip"},
			{InDst: "", InSrc: "./x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "", InSrc: "long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
			{InDst: "", InSrc: "/long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutSrc: "/long/path/to/x.zip"},
			{InDst: "", InSrc: "./long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},

			{InDst: "y.zip", InSrc: "x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "y.zip", InSrc: "/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "/x.zip"},
			{InDst: "y.zip", InSrc: "./x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "y.zip", InSrc: "long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
			{InDst: "y.zip", InSrc: "/long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "/long/path/to/x.zip"},
			{InDst: "y.zip", InSrc: "./long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},

			{InDst: "/y.zip", InSrc: "x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "/y.zip", InSrc: "/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "/x.zip"},
			{InDst: "/y.zip", InSrc: "./x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "/y.zip", InSrc: "long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
			{InDst: "/y.zip", InSrc: "/long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "/long/path/to/x.zip"},
			{InDst: "/y.zip", InSrc: "./long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},

			{InDst: "./y.zip", InSrc: "x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "./y.zip", InSrc: "/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "/x.zip"},
			{InDst: "./y.zip", InSrc: "./x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "./y.zip", InSrc: "long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
			{InDst: "./y.zip", InSrc: "/long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "/long/path/to/x.zip"},
			{InDst: "./y.zip", InSrc: "./long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},

			{InDst: "long/path/to/y.zip", InSrc: "x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "/x.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "./x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "/long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "/long/path/to/x.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "./long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},

			{InDst: "/long/path/to/y.zip", InSrc: "x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "/x.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "./x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "/long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "/long/path/to/x.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "./long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},

			{InDst: "./long/path/to/y.zip", InSrc: "x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "/x.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "./x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "x.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "/long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "/long/path/to/x.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "./long/path/to/x.zip", OutDst: fmt.Sprintf("artifacts/%s/1/long/path/to/y.zip", resolver.ResourceTypePlural), OutSrc: "long/path/to/x.zip"},
		}

		for _, assertion := range assertions {
			paths, err := resolver.Resolve(OperationPush, assertion.InSrc, assertion.InDst)
			assert.Nil(t, err)
			assert.Equal(t, assertion.OutSrc, paths.Source)
			assert.Equal(t, assertion.OutDst, paths.Destination)
		}
	})

	t.Run(testCase.ResourceType+" resolve paths for pull", func(t *testing.T) {
		os.Setenv(testCase.EnvironmentVariable, "1")
		resolver, _ := NewPathResolver(testCase.ResourceType, "")

		assertions := []resolvedPathAssertion{
			{InDst: "", InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "x.zip"},
			{InDst: "", InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "x.zip"},
			{InDst: "", InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "x.zip"},
			{InDst: "", InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "x.zip"},
			{InDst: "", InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "x.zip"},
			{InDst: "", InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "x.zip"},

			{InDst: "y.zip", InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "y.zip", InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "y.zip", InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "y.zip", InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "y.zip", InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "y.zip", InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},

			{InDst: "/y.zip", InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "/y.zip"},
			{InDst: "/y.zip", InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "/y.zip"},
			{InDst: "/y.zip", InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "/y.zip"},
			{InDst: "/y.zip", InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "/y.zip"},
			{InDst: "/y.zip", InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "/y.zip"},
			{InDst: "/y.zip", InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "/y.zip"},

			{InDst: "./y.zip", InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "./y.zip", InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "./y.zip", InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "./y.zip", InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "./y.zip", InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},
			{InDst: "./y.zip", InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "y.zip"},

			{InDst: "long/path/to/y.zip", InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "long/path/to/y.zip", InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},

			{InDst: "/long/path/to/y.zip", InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "/long/path/to/y.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "/long/path/to/y.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "/long/path/to/y.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "/long/path/to/y.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "/long/path/to/y.zip"},
			{InDst: "/long/path/to/y.zip", InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "/long/path/to/y.zip"},

			{InDst: "./long/path/to/y.zip", InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
			{InDst: "./long/path/to/y.zip", InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural), OutDst: "long/path/to/y.zip"},
		}

		for _, assertion := range assertions {
			paths, err := resolver.Resolve(OperationPull, assertion.InSrc, assertion.InDst)
			assert.Nil(t, err)
			assert.Equal(t, assertion.OutSrc, paths.Source)
			assert.Equal(t, assertion.OutDst, paths.Destination)
		}
	})

	t.Run(testCase.ResourceType+" resolve paths for yank", func(t *testing.T) {
		os.Setenv(testCase.EnvironmentVariable, "1")
		resolver, _ := NewPathResolver(testCase.ResourceType, "")

		assertions := []resolvedPathAssertion{
			{InSrc: "x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural)},
			{InSrc: "/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural)},
			{InSrc: "./x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/x.zip", resolver.ResourceTypePlural)},
			{InSrc: "long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural)},
			{InSrc: "/long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural)},
			{InSrc: "./long/path/to/x.zip", OutSrc: fmt.Sprintf("artifacts/%s/1/long/path/to/x.zip", resolver.ResourceTypePlural)},
		}

		for _, assertion := range assertions {
			paths, err := resolver.Resolve(OperationYank, assertion.InSrc, "")
			assert.Nil(t, err)
			assert.Equal(t, assertion.OutSrc, paths.Source)
			assert.Empty(t, assertion.OutDst)
		}
	})
}
