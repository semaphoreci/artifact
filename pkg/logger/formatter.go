package logger

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type CustomFormatter struct {
}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	log := fmt.Sprintf("[%-19s] %s", entry.Time.UTC().Format(time.StampMilli), entry.Message)
	return []byte(log), nil
}
