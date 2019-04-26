package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ParseRelativeAgeForHumans converts a human readable time duration to a machine readable one.
// Supported options are:
//  empty: -1 means NaN
//  Nd for N days
//  Nw for N weeks
//  Nm for N months
//  Ny for N years
func ParseRelativeAgeForHumans(descriptor string) (time.Duration, error) {
	descLen := len(descriptor)
	descriptor = strings.ToLower(descriptor)
	if descLen == 0 || descriptor == "never" {
		return -1, nil // never expires
	}

	num, err := strconv.Atoi(descriptor[:descLen-1])
	if err != nil {
		return 0, fmt.Errorf("Failed to parse time for humans: %s, error: %s", descriptor, err)
	}
	numDur := time.Duration(num) * time.Hour * 24
	lastByte := descriptor[descLen-1]
	lastRune := unicode.ToLower(rune(lastByte))
	if lastRune == 'd' {
		return numDur, nil
	}
	if lastRune == 'w' {
		return numDur * 7, nil
	}
	if lastRune == 'm' {
		return numDur * 30, nil
	}
	if lastRune == 'y' {
		return numDur * 365, nil
	}
	return 0, fmt.Errorf("Failed to parse time for humans: %s, error: %s", descriptor, err)
}
