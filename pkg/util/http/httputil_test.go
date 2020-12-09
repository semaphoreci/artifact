package httputil

import (
	"net/http"
	"testing"
)

func TestIsStatusOK(t *testing.T) {
	check := func(statusCode int, exp bool) {
		ok := IsStatusOK(statusCode)
		if ok != exp {
			t.Errorf("IsStatusOk doesn't match %t != expected(%t)", ok, exp)
		}
	}

	check(http.StatusContinue, false)
	check(http.StatusProcessing, false)
	check(http.StatusOK, true)
	check(http.StatusAccepted, true)
	check(http.StatusMultiStatus, true)
	check(http.StatusProcessing, false)
	check(http.StatusMovedPermanently, false)
	check(http.StatusSeeOther, false)
	check(http.StatusNotModified, false)
	check(http.StatusBadRequest, false)
	check(http.StatusUnauthorized, false)
	check(http.StatusInternalServerError, false)
	check(http.StatusNotImplemented, false)
	check(http.StatusBadGateway, false)
}
