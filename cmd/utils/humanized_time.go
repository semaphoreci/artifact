package utils

import (
	"fmt"
	"strconv"
	"time"
	"unicode"
)

// ParseRelativeAgeForHumans converts a human readable time duration to a machine readable one.
// Supported options are:
//  empty: -1 means NaN
//  just integer (number of seconds)
//  Nh for N hours
//  Nd for N days
//  Nw for N weeks
//  Nm for N months
//  Ny for N years
func ParseRelativeAgeForHumans(descriptor string) (time.Duration, error) {
	descLen := len(descriptor)
	if descLen == 0 {
		return -1, nil // never expires
	}

	lastByte := descriptor[descLen-1]
	if lastByte >= 48 && lastByte <= 57 { // probably numbers only
		sec, err := strconv.Atoi(descriptor)
		if err != nil {
			return 0, fmt.Errorf("Failed to parse time for humans: %s, error: %s", descriptor, err)
		}
		return time.Duration(sec) * time.Second, nil
	}
	num, err := strconv.Atoi(descriptor[:descLen-1])
	if err != nil {
		return 0, fmt.Errorf("Failed to parse time for humans: %s, error: %s", descriptor, err)
	}
	numDur := time.Duration(num) * time.Hour
	lastRune := unicode.ToLower(rune(lastByte))
	if lastRune == 'h' {
		return numDur, nil
	}
	if lastRune == 'd' {
		return numDur * 24, nil
	}
	if lastRune == 'w' {
		return numDur * 168, nil // 24 * 7
	}
	if lastRune == 'm' {
		return numDur * 720, nil // 24 * 30
	}
	if lastRune == 'y' {
		return numDur * 8760, nil // 24 * 365
	}
	return 0, fmt.Errorf("Failed to parse time for humans: %s, error: %s", descriptor, err)
}
