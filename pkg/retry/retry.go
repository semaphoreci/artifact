package retry

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func RetryWithConstantWait(task string, maxAttempts int, wait time.Duration, f func() error) error {
	for attempt := 1; ; attempt++ {
		err := f()
		if err == nil {
			return nil
		}

		if !strings.Contains(err.Error(), "500") {
			return err
		}
		if attempt >= maxAttempts {
			return fmt.Errorf("[%s] failed after [%d] attempts - giving up: %v", task, attempt, err)
		}

		log.Errorf("[%s] attempt [%d] failed with [%v] - retrying in %s...\n", task, attempt, err, wait)
		time.Sleep(wait)
	}
}
