package errutil

import (
	"fmt"
	"time"

	"github.com/semaphoreci/artifact/pkg/util/log"
	"go.uber.org/zap"
)

const ( // TODO: these constants may be moved to conf or console arg
	// RetryLimit is a number how many times bucket creation is retried before returning an error.
	RetryLimit = 3
	// startTimeout is the starting timeout for requests in milliseconds.
	startTimeout = time.Duration(2500)
	// addTimeout is the the amount, that is added to timeout for each retry.
	addTimeout = time.Duration(500)
)

// RetryOnFailure calls the given function for a certain (RetryLimit) number of times. The
// function should be an inline function, so it can set return values. The function returns an
// error. The retries stop, if the result is ok. After the certain number of times expired, it
// logs the error anyway.
func RetryOnFailure(msg string, toRun func() bool) (ok bool) {
	timeout := startTimeout
	for i := 0; i < RetryLimit; i++ {
		if ok = toRun(); ok {
			return
		}
		if i == 0 {
			log.VerboseError(fmt.Sprintf("Failed to %s, retrying...", msg))
		}
		time.Sleep(timeout * time.Millisecond)
		timeout += addTimeout
	}
	log.VerboseError("Repeatedly failed to "+msg, zap.Int("retry number", RetryLimit))
	return
}
