// +build !windows

package termstatus

import "io"

// clearLines will clear the current line and the n lines above. Afterwards the
// cursor is positioned at the start of the first cleared line.
func clearLines(wr io.Writer, n int) error {
	return posixClearLines(wr, n)
}
