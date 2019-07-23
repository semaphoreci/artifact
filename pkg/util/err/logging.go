package errutil

import (
	"fmt"
	"log"

	"github.com/semaphoreci/artifact/config"
)

var (
	// Debug is the logger function for the debug level.
	Debug func(string, ...interface{})
	// Info is the logger function for the info level.
	Info func(string, ...interface{})
	// Warn is the logger function for the warning level.
	Warn func(string, error) error
	// Error is the logger function for the error level.
	Error func(string, error) error
)

func init() {
	logLvl := config.LogLevel
	Error = logStatus
	Warn = noopStatus
	Info = noop
	Debug = noop
	if logLvl < config.LogLvlError {
		Warn = logStatus
		if logLvl < config.LogLvlWarn {
			Info = logStr
			if logLvl < config.LogLvlInfo {
				Debug = logStr
			}
		}
	}
}

func noop(string, ...interface{}) {}

func logStr(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func noopStatus(msg string, err error) error {
	return nil
}

// logStatus creates a new error with grpc status code, logs and returns it.
func logStatus(msg string, err error) error {
	err = fmt.Errorf("%s: %v", msg, err)
	log.Println(err)
	return err
}
