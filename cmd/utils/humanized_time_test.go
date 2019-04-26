package utils

import (
	"testing"
	"time"
)

func TestParseRelativeAgeForHumans(t *testing.T) {
	testParseRelativeAgeForHumans := func(descriptor string, duration time.Duration, isErr bool) {
		d, e := ParseRelativeAgeForHumans(descriptor)
		if (e != nil) != isErr {
			t.Errorf("Error state differs for descriptor %s: %t != expected(%t)", descriptor, e != nil, isErr)
			return
		}
		if e != nil { // if there's an error, the value should be zero
			if d != 0 {
				t.Errorf("Erroneous result must be zero for descriptor %s: %d != 0", descriptor, d)
			}
			return
		}
		if d != duration {
			t.Errorf("Valid result doesn't match for descriptor %s: %d != expected(%d)", descriptor, d, duration)
		}
	}

	day := 24 * time.Hour
	testParseRelativeAgeForHumans("", -1, false)
	testParseRelativeAgeForHumans("never", -1, false)
	testParseRelativeAgeForHumans("Never", -1, false)
	testParseRelativeAgeForHumans("0", 0, true)
	testParseRelativeAgeForHumans("1", 0, true)
	testParseRelativeAgeForHumans("10", 0, true)
	testParseRelativeAgeForHumans("1000", 0, true)
	testParseRelativeAgeForHumans("h0", 0, true)
	testParseRelativeAgeForHumans("1h0", 0, true)
	testParseRelativeAgeForHumans("h", 0, true)
	testParseRelativeAgeForHumans("hw", 0, true)
	testParseRelativeAgeForHumans("1s", 0, true)
	testParseRelativeAgeForHumans("1a", 0, true)
	testParseRelativeAgeForHumans("1h", 0, true)
	testParseRelativeAgeForHumans("10h", 0, true)
	testParseRelativeAgeForHumans("1000h", 0, true)
	testParseRelativeAgeForHumans("1d", day, false)
	testParseRelativeAgeForHumans("10D", 10*day, false)
	testParseRelativeAgeForHumans("1000d", 1000*day, false)
	testParseRelativeAgeForHumans("1w", 7*day, false)
	testParseRelativeAgeForHumans("10w", 70*day, false)
	testParseRelativeAgeForHumans("1000W", 7000*day, false)
	testParseRelativeAgeForHumans("1M", 30*day, false)
	testParseRelativeAgeForHumans("10m", 300*day, false)
	testParseRelativeAgeForHumans("1000m", 30000*day, false)
	testParseRelativeAgeForHumans("1y", 365*day, false)
	testParseRelativeAgeForHumans("10Y", 3650*day, false)
	testParseRelativeAgeForHumans("100y", 36500*day, false)
}
