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

	testParseRelativeAgeForHumans("", -1, false)
	testParseRelativeAgeForHumans("0", 0, false)
	testParseRelativeAgeForHumans("1", time.Second, false)
	testParseRelativeAgeForHumans("10", 10*time.Second, false)
	testParseRelativeAgeForHumans("1000", 1000*time.Second, false)
	testParseRelativeAgeForHumans("h0", 0, true)
	testParseRelativeAgeForHumans("1h0", 0, true)
	testParseRelativeAgeForHumans("h", 0, true)
	testParseRelativeAgeForHumans("hw", 0, true)
	testParseRelativeAgeForHumans("1s", 0, true)
	testParseRelativeAgeForHumans("1a", 0, true)
	testParseRelativeAgeForHumans("1h", time.Hour, false)
	testParseRelativeAgeForHumans("10h", 10*time.Hour, false)
	testParseRelativeAgeForHumans("1000h", 1000*time.Hour, false)
	testParseRelativeAgeForHumans("1d", 24*time.Hour, false)
	testParseRelativeAgeForHumans("10d", 240*time.Hour, false)
	testParseRelativeAgeForHumans("1000d", 24000*time.Hour, false)
	testParseRelativeAgeForHumans("1w", 168*time.Hour, false)
	testParseRelativeAgeForHumans("10w", 1680*time.Hour, false)
	testParseRelativeAgeForHumans("1000w", 168000*time.Hour, false)
	testParseRelativeAgeForHumans("1m", 720*time.Hour, false)
	testParseRelativeAgeForHumans("10m", 7200*time.Hour, false)
	testParseRelativeAgeForHumans("1000m", 720000*time.Hour, false)
	testParseRelativeAgeForHumans("1y", 8760*time.Hour, false)
	testParseRelativeAgeForHumans("10y", 87600*time.Hour, false)
	testParseRelativeAgeForHumans("100y", 876000*time.Hour, false)
}
