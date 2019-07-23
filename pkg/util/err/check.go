package errutil

import (
	"flag"
	"fmt"
	"os"
)

// Check checks if an error is present.
// If it is present, it displays the error and exits with status 1.
// If you want to display a custom message use CheckWithMessage.
func Check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())

		Exit(1)
	}
}

// Exit quits the application with a given value.
func Exit(code int) {
	if flag.Lookup("test.v") == nil {
		os.Exit(1)
	} else {
		panic(fmt.Sprintf("exit %d", code))
	}
}
