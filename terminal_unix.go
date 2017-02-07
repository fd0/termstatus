// +build !windows

package termstatus

import isatty "github.com/mattn/go-isatty"

// clearLines will clear the current line and the n lines above. Afterwards the
// cursor is positioned at the start of the first cleared line.
func clearLines(wr TerminalWriter) func(TerminalWriter, int) error {
	return posixClearLines
}

// canUpdateStatus returns true if status lines can be printed, the process
// output is not redirected to a file or pipe.
func canUpdateStatus(wr TerminalWriter) bool {
	return isatty.IsTerminal(wr.Fd())
}
