package internal

import (
	"errors"
	"fmt"
)

// ErrAlreadyExists signifies, that a file or directory can't be overriden, because it already exists.
type ErrAlreadyExists struct {
	Filename string
	Location string
}

// Error returns the whole error about the existing file.
func (err *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("The file '%s' already exists in the %s", err.Filename, err.Location)
}

// ErrUnknownGCS is returned for a query on Google Cloud Storage, resulting an unknown error.
type ErrUnknownGCS struct {
	ErrStr string
}

// Error returns the original error, wrapped in a describing text.
func (err *ErrUnknownGCS) Error() string {
	return fmt.Sprintf("Unknown error happened in the Google Cloud Storage: %s", err.ErrStr)
}

// ErrDirectoryFoundGCS is not really an error. It means that a dir is found on the Google Cloud Storage.
var ErrDirectoryFoundGCS = errors.New("ErrDirectoryFoundGCS")

// ErrNotFound is returned when a source copied from doesn't exists.
type ErrNotFound ErrAlreadyExists

// Error returns the whole error about the existing file.
func (err *ErrNotFound) Error() string {
	return fmt.Sprintf("The file or directory '%s' doesn't exists in the %s", err.Filename, err.Location)
}
