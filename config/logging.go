package config

import (
	"fmt"
	"log"
	"os"
)

// LogLevel is the global logging level for the application.

var LogLevel LogLevels

// LogLevels contains available logging levels.
type LogLevels int

const (
	// LogLvlDebug is the debug log level.
	LogLvlDebug LogLevels = iota
	// LogLvlInfo is the info log level.
	LogLvlInfo
	// LogLvlWarn is the warning log level.
	LogLvlWarn
	// LogLvlError is the error log level.
	LogLvlError
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	logLvlStrToVal := map[string]LogLevels{
		"DEBUG": LogLvlDebug,
		"INFO":  LogLvlInfo,
		"WARN":  LogLvlWarn,
		"ERROR": LogLvlError,
		"":      LogLvlError, // the default level is error
	}
	var ok bool
	if LogLevel, ok = logLvlStrToVal[os.Getenv("ARTIFACT_LOGLEVEL")]; !ok {
		panic(fmt.Errorf("ARTIFACT_LOGLEVEL must be ['DEBUG', 'INFO', 'WARN', 'ERROR'] or empty"))
	}
}
