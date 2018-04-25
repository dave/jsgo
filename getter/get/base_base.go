package get

import (
	"bytes"
	"errors"
	"go/scanner"
)

// ExpandScanner expands a scanner.List error into all the errors in the list.
// The default Error method only shows the first error
// and does not shorten paths.
func ExpandScanner(err error) error {
	// Look for parser errors.
	if err, ok := err.(scanner.ErrorList); ok {
		// Prepare error with \n before each message.
		// When printed in something like context: %v
		// this will put the leading file positions each on
		// its own line. It will also show all the errors
		// instead of just the first, as err.Error does.
		var buf bytes.Buffer
		for _, e := range err {
			buf.WriteString("\n")
			buf.WriteString(e.Error())
		}
		return errors.New(buf.String())
	}
	return err
}
