//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"fmt"
	"io"
	"os"
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
func openInput(inFile string) (io.ReadCloser, func(), error) {
	var in io.ReadCloser

	if inFile != "" {
		f, err := os.Open(inFile)
		if err != nil {
			return nil, nil,
				fmt.Errorf("failed to open input file: %w", err)
		}
		in = f
	} else {
		in = os.Stdin
	}

	cleanup := func() {
		if in != os.Stdin {
			err := in.Close()
			if err != nil {
				fmt.Printf("Failed to close input file: %s\n", err.Error())
			}
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
//   - error: File creation errors
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
func openOutput(outFile string) (io.Writer, func(), error) {
	var out io.Writer
	var outCloser io.Closer

	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return nil, nil,
				fmt.Errorf("failed to create output file: %w", err)
		}
		out = f
		outCloser = f
	} else {
		out = os.Stdout
	}

	cleanup := func() {
		if outCloser != nil {
			err := outCloser.Close()
			if err != nil {
				fmt.Printf("Failed to close output file: %s\n",
					err.Error())
			}
		}
	}

	return out, cleanup, nil
}
