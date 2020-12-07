package errutil

import (
	"github.com/semaphoreci/artifact/pkg/util/log"
	"go.uber.org/zap"
)

const (
	lfs = "local file system"
	gcs = "Google Cloud Storage"
)

// Location is an enum for choosing between local file system of Google Cloud Storage.
type Location int

const (
	// Lfs is the local file system enum value.
	Lfs Location = iota
	// Gcs is the Google Cloud Storage enum value.
	Gcs
)

var lfsMap = map[Location]string{
	Lfs: lfs,
	Gcs: gcs,
}

// ErrAlreadyExists is called in case of a file conflict.
func ErrAlreadyExists(description, filename string, location Location) {
	log.Error("The file already exists; delete it first, or use --force flag",
		zap.String("name", filename), zap.String("while", description),
		zap.String("location", lfsMap[location]))
}

// ErrNotFound is called when a source copied from doesn't exists.
func ErrNotFound(filename string, location Location) {
	log.Error("The file or directory doesn't exists", zap.String("name", filename),
		zap.String("location", lfsMap[location]))
}
