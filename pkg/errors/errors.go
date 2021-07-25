// Package errors collects helper functions for dealing with errors.
package errors

import (
	"fmt"
)

var (
	// ErrorInvalidValue is the error returned when an invalid value is
	// encountered.
	ErrorInvalidValue = fmt.Errorf("invalid value specified")

	// ErrorOperationFailed is the error returned when an operation execution
	// encounters an error.
	ErrorOperationFailed = fmt.Errorf("executing operation failed")
)

// NewErrorWithDetails adds structured context to an error. In case of a nil error
// specified or no key value pairs provided, the original value is returned let
// it be either nil or an error. In case annotations do happen, the original
// error is left intact and a different annotated error value is returned.
func NewErrorWithDetails(originalError error, keyValuePairs ...interface{}) (err error) {
	if originalError == nil ||
		len(keyValuePairs) == 0 {
		return originalError
	}

	if len(keyValuePairs)%2 == 1 {
		keyValuePairs = append(keyValuePairs, "")
	}

	err = fmt.Errorf("%w", originalError)
	for argumentIndex := 0; argumentIndex < len(keyValuePairs)-1; argumentIndex += 2 {
		err = fmt.Errorf("%w, %+v: '%+v'", err, keyValuePairs[argumentIndex], keyValuePairs[argumentIndex+1])
	}

	return err
}
