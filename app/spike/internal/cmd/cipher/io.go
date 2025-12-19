//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"fmt"
	"io"
	"os"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// openInput opens a file for reading or returns stdin if no file is
// specified. The returned closer should be called to clean up resources.
//
// Parameters:
//   - inFile: Input file path (empty string means stdin)
//
// Returns:
//   - io.ReadCloser: Reader for the input source
//   - func(): Cleanup function to close the file (safe to call always)
//   - error: File opening errors
//
// The cleanup function is safe to call even if stdin is used (it will only
// close actual files, not stdin).
//
// Example usage:
//
//	in, cleanup, err := openInput(inFile)
//	if err != nil {
//	    return err
//	}
//	defer cleanup()
func openInput(inFile string) (io.ReadCloser, func(), *sdkErrors.SDKError) {
	var in io.ReadCloser

	if inFile != "" {
		f, err := os.Open(inFile)
		if err != nil {
			failErr := sdkErrors.ErrFSFileOpenFailed.Wrap(err)
			failErr.Msg = fmt.Sprintf("failed to open input file: %s", inFile)
			return nil, func() {}, failErr
		}
		in = f
	} else {
		in = os.Stdin
	}

	cleanup := func() {
		if in != os.Stdin {
			// Best-effort close; nothing actionable if it fails.
			_ = in.Close()
		}
	}

	return in, cleanup, nil
}

// openOutput creates a file for writing or returns stdout if no file is
// specified. The returned closer should be called to clean up resources.
//
// Parameters:
//   - outFile: Output file path (empty string means stdout)
//
// Returns:
//   - io.Writer: Writer for the output destination
//   - func(): Cleanup function to close the file (safe to call always)
//   - *sdkErrors.SDKError: File creation errors
//
// The cleanup function is safe to call even if stdout is used (it will only
// close actual files, not stdout).
//
// Example usage:
//
//	out, cleanup, err := openOutput(outFile)
//	if err != nil {
//	    return err
//	}
//	defer cleanup()
func openOutput(outFile string) (io.Writer, func(), *sdkErrors.SDKError) {
	var out io.Writer
	var outCloser io.Closer

	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			// Using ErrFSFileOpenFailed for file creation errors as well,
			// since it represents file access failures in general
			failErr := sdkErrors.ErrFSFileOpenFailed.Wrap(err)
			failErr.Msg = fmt.Sprintf("failed to create output file: %s", outFile)
			return nil, func() {}, failErr
		}
		out = f
		outCloser = f
	} else {
		out = os.Stdout
	}

	cleanup := func() {
		if outCloser != nil {
			// Best-effort close; nothing actionable if it fails.
			_ = outCloser.Close()
		}
	}

	return out, cleanup, nil
}
