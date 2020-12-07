package errutil

import (
	"testing"
)

func TestRetrySuccess(t *testing.T) {
	testVal := 2
	failTwice := func() (ok bool) {
		defer func() { testVal-- }()
		return testVal <= 0
	}

	if ok := RetryOnFailure("get mock result", failTwice); !ok {
		t.Errorf("Should be success")
	}
}

func TestRetryableFailure(t *testing.T) {
	testVal := 3
	failThreetimes := func() (ok bool) {
		defer func() { testVal-- }()
		return testVal <= 0
	}

	if ok := RetryOnFailure("get mock result", failThreetimes); ok {
		t.Errorf("Should be failure")
	}
}
