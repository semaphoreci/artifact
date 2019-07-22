package errutil

import (
	"fmt"
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

// ErrAlreadyExists signifies, that a file or directory can't be overriden, because it already exists.
type ErrAlreadyExists struct {
	Filename string
	Location Location
}

// Error returns the whole error about the existing file.
func (err *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("The file '%s' already exists in the %s. It can be overwritten with --force flag",
		err.Filename, lfsMap[err.Location])
}

// ErrNotFound is returned when a source copied from doesn't exists.
type ErrNotFound ErrAlreadyExists

// Error returns the whole error about the existing file.
func (err *ErrNotFound) Error() string {
	return fmt.Sprintf("The file or directory '%s' doesn't exists in the %s",
		err.Filename, lfsMap[err.Location])
}
