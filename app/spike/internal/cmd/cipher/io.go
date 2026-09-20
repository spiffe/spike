//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// closeOnce returns a cleanup function that closes the given closer the
// first time it is called and reports the close error. Later calls are
// no-ops that return nil, so callers may invoke the cleanup more than once.
// A nil closer yields a cleanup that never does anything; this is how stdin
// and stdout are protected from being closed.
//
// Parameters:
//   - c: The closer to close, or nil for a cleanup that does nothing
//   - sentinel: The SDK error to wrap a failed close in
//   - what: A short description of the resource, used in the error message
//
// Returns:
//   - func() *sdkErrors.SDKError: The cleanup function; it returns the
//     wrapped close error on the first call if the close fails, and nil
//     otherwise
func closeOnce(
	c io.Closer, sentinel *sdkErrors.SDKError, what string,
) func() *sdkErrors.SDKError {
	closed := false
	return func() *sdkErrors.SDKError {
		if c == nil || closed {
			return nil
		}
		closed = true
		if err := c.Close(); err != nil {
			failErr := sentinel.Wrap(err)
			failErr.Msg = fmt.Sprintf("failed to close %s", what)
			return failErr
		}
		return nil
	}
}

// openInput opens a file for reading or returns stdin if no file is
// specified. The returned cleanup function must be called to release the
// file, and its error must be handled.
//
// Parameters:
//   - inFile: Input file path (empty string means stdin)
//
// Returns:
//   - io.ReadCloser: Reader for the input source
//   - func() *sdkErrors.SDKError: Cleanup function that closes the file
//     (safe to call more than once; it never closes stdin)
//   - *sdkErrors.SDKError: File opening errors
//
// Example usage:
//
//	in, cleanup, err := openInput(inFile)
//	if err != nil {
//	    return err
//	}
//	defer func() {
//	    if closeErr := cleanup(); closeErr != nil && err == nil {
//	        err = closeErr
//	    }
//	}()
func openInput(
	inFile string,
) (io.ReadCloser, func() *sdkErrors.SDKError, *sdkErrors.SDKError) {
	if inFile == "" {
		return os.Stdin, closeOnce(nil, nil, ""), nil
	}

	// The path is supplied by the operator on the command line; cleaning
	// it normalizes separators and removes redundant elements.
	f, err := os.Open(filepath.Clean(inFile))
	if err != nil {
		failErr := sdkErrors.ErrFSFileOpenFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf("failed to open input file: %s", inFile)
		return nil, closeOnce(nil, nil, ""), failErr
	}

	return f, closeOnce(
		f, sdkErrors.ErrFSFileCloseFailed, "input file "+inFile,
	), nil
}

// openOutput creates a file for writing or returns stdout if no file is
// specified. The returned cleanup function must be called to release the
// file, and its error must be handled: a failed close on an output file can
// mean that not all of the written data reached the disk.
//
// Parameters:
//   - outFile: Output file path (empty string means stdout)
//
// Returns:
//   - io.Writer: Writer for the output destination
//   - func() *sdkErrors.SDKError: Cleanup function that closes the file
//     (safe to call more than once; it never closes stdout)
//   - *sdkErrors.SDKError: File creation errors
//
// Example usage:
//
//	out, cleanup, err := openOutput(outFile)
//	if err != nil {
//	    return err
//	}
//	defer func() {
//	    if closeErr := cleanup(); closeErr != nil && err == nil {
//	        err = closeErr
//	    }
//	}()
func openOutput(
	outFile string,
) (io.Writer, func() *sdkErrors.SDKError, *sdkErrors.SDKError) {
	if outFile == "" {
		return os.Stdout, closeOnce(nil, nil, ""), nil
	}

	// The path is supplied by the operator on the command line; cleaning
	// it normalizes separators and removes redundant elements.
	f, err := os.Create(filepath.Clean(outFile))
	if err != nil {
		// ErrFSFileOpenFailed covers file creation as well, since it
		// represents file access failures in general.
		failErr := sdkErrors.ErrFSFileOpenFailed.Wrap(err)
		failErr.Msg = fmt.Sprintf("failed to create output file: %s", outFile)
		return nil, closeOnce(nil, nil, ""), failErr
	}

	return f, closeOnce(
		f, sdkErrors.ErrFSFileCloseFailed, "output file "+outFile,
	), nil
}
