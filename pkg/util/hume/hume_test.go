package humeutil

import (
	"testing"
	"time"
)

func TestParseRelativeAgeForHumans(t *testing.T) {
	testParseRelativeAgeForHumans := func(descriptor string, duration time.Duration) {
		d := ParseRelativeAgeForHumans(descriptor)
		if d != duration {
			t.Errorf("Valid result doesn't match for descriptor %s: %d != expected(%d)",
				descriptor, d, duration)
		}
	}

	day := 24 * time.Hour
	testParseRelativeAgeForHumans("", -1)
	testParseRelativeAgeForHumans("never", -1)
	testParseRelativeAgeForHumans("Never", -1)
	testParseRelativeAgeForHumans("0", 0)
	testParseRelativeAgeForHumans("1", 0)
	testParseRelativeAgeForHumans("10", 0)
	testParseRelativeAgeForHumans("1000", 0)
	testParseRelativeAgeForHumans("h0", 0)
	testParseRelativeAgeForHumans("1h0", 0)
	testParseRelativeAgeForHumans("h", 0)
	testParseRelativeAgeForHumans("hw", 0)
	testParseRelativeAgeForHumans("1s", 0)
	testParseRelativeAgeForHumans("1a", 0)
	testParseRelativeAgeForHumans("1h", 0)
	testParseRelativeAgeForHumans("10h", 0)
	testParseRelativeAgeForHumans("1000h", 0)
	testParseRelativeAgeForHumans("1d", day)
	testParseRelativeAgeForHumans("10D", 10*day)
	testParseRelativeAgeForHumans("1000d", 1000*day)
	testParseRelativeAgeForHumans("1w", 7*day)
	testParseRelativeAgeForHumans("10w", 70*day)
	testParseRelativeAgeForHumans("1000W", 7000*day)
	testParseRelativeAgeForHumans("1M", 30*day)
	testParseRelativeAgeForHumans("10m", 300*day)
	testParseRelativeAgeForHumans("1000m", 30000*day)
	testParseRelativeAgeForHumans("1y", 365*day)
	testParseRelativeAgeForHumans("10Y", 3650*day)
	testParseRelativeAgeForHumans("100y", 36500*day)
}
