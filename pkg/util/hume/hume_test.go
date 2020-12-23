package humeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRelativeAgeForHumans(t *testing.T) {
	check := func(descriptor string, expDuration time.Duration) {
		d := ParseRelativeAgeForHumans(descriptor)
		assert.Equal(t, expDuration, d, descriptor)
	}

	day := 24 * time.Hour
	check("", -1)
	check("never", -1)
	check("Never", -1)
	check("0", 0)
	check("1", 0)
	check("10", 0)
	check("1000", 0)
	check("h0", 0)
	check("1h0", 0)
	check("h", 0)
	check("hw", 0)
	check("1s", 0)
	check("1a", 0)
	check("1h", 0)
	check("10h", 0)
	check("1000h", 0)
	check("1d", day)
	check("10D", 10*day)
	check("1000d", 1000*day)
	check("1w", 7*day)
	check("10w", 70*day)
	check("1000W", 7000*day)
	check("1M", 30*day)
	check("10m", 300*day)
	check("1000m", 30000*day)
	check("1y", 365*day)
	check("10Y", 3650*day)
	check("100y", 36500*day)
}
