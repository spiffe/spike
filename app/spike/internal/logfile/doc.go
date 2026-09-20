//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package logfile binds the SDK logger to a well-known file so that SPIKE
// Pilot never writes structured log lines to stdout.
//
// SPIKE Pilot writes command output to stdout and human-readable errors to
// stderr. The SDK logger, however, binds to os.Stdout the first time it is
// used and offers no way to select another writer. A fatal log line on
// stdout would corrupt redirected command output, for example a secret
// written to a file. This package gives the logger a file instead:
// $HOME/.spike/pilot.log, or /tmp/.spike-$USER/pilot.log when no home
// directory is available. The resolution order mirrors the SDK's recovery
// directory so that all Pilot state lives side by side.
package logfile
