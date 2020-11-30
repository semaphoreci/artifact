package humeutil

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"

	errutil "github.com/semaphoreci/artifact/pkg/util/err"
)

// ParseRelativeAgeForHumans converts a human readable time duration to a machine readable one.
// Supported options are:
//  empty: -1 means no expiration
//  Nd for N days
//  Nw for N weeks
//  Nm for N months
//  Ny for N years
// Returned 0 means an error that is already logged.
func ParseRelativeAgeForHumans(descriptor string) time.Duration {
	descLen := len(descriptor)
	descriptor = strings.ToLower(descriptor)
	if descLen == 0 || descriptor == "never" {
		return -1 // never expires
	}

	num, err := strconv.Atoi(descriptor[:descLen-1])
	if err != nil {
		errutil.L.Error("parsing time failed", zap.String("time to parse", descriptor),
			zap.Error(err))
		return 0
	}
	numDur := time.Duration(num) * time.Hour * 24
	lastByte := descriptor[descLen-1]
	lastRune := unicode.ToLower(rune(lastByte))
	if lastRune == 'd' {
		return numDur
	}
	if lastRune == 'w' {
		return numDur * 7
	}
	if lastRune == 'm' {
		return numDur * 30
	}
	if lastRune == 'y' {
		return numDur * 365
	}
	errutil.L.Error("parsing time failed", zap.String("time to parse", descriptor))
	return 0
}
