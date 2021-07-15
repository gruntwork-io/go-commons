// This file contains modified versions of functions vendored in the golang standard library, including private
// functions that are used by that function.
package shell

import "bytes"

// ScanLinesIncludeRaw is a modified version of bufio.ScanLines that returns the newlines when scanning, unless it hits
// the EOF. This is necessary so that we can return an accurate representation of what was outputted in the shell
// (e.g., if the shell does NOT contain a newline at the end, it should be omitted - similarly, if the shell contains a
// newline at the end, it should be included).
// bufio.ScanLines is licensed under a BSD-style license.
// https://github.com/golang/go/blob/93200b98c75500b80a2bf7cc31c2a72deff2741c/src/bufio/scan.go#L345
func ScanLinesIncludeRaw(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line, but make sure to append the newline token before returning.
		return i + 1, append(dropCR(data[0:i]), '\n'), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data. This is the same implementation as bufio.dropCR.
// bufio.dropCR function is licensed under a BSD-style license.
// https://github.com/golang/go/blob/93200b98c75500b80a2bf7cc31c2a72deff2741c/src/bufio/scan.go#L337
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
