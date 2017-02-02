// +build windows

package termstatus

import (
	"os"
	"syscall"
	"unsafe"

	isatty "github.com/mattn/go-isatty"
)

// clearLines clears the current line and n lines above it.
func clearLines(wr Terminal, n int) error {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return posixClearLines(wr, n)
	}

	return windowsClearLines(wr, n)
}

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
)

type (
	short int16
	word  uint16
	dword uint32

	coord struct {
		x short
		y short
	}
	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word
		window            smallRect
		maximumWindowSize coord
	}
)

// windowsClearLines clears the current line and n lines above it.
func windowsClearLines(wr Terminal, n int) error {
	var info consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(wr.Fd(), uintptr(unsafe.Pointer(&info)))

	for i := 0; i <= n; i++ {
		// clear the line
		cursor := coord{
			x: info.window.left,
			y: info.cursorPosition.y - short(i),
		}
		var count, w dword
		count = dword(info.size.x)
		procFillConsoleOutputCharacter.Call(wr.Fd(), uintptr(' '), uintptr(count), *(*uintptr)(unsafe.Pointer(&cursor)), uintptr(unsafe.Pointer(&w)))
	}

	// move cursor up by n lines and to the first column
	info.cursorPosition.y -= short(n)
	info.cursorPosition.x = 0
	procSetConsoleCursorPosition.Call(wr.Fd(), uintptr(*(*int32)(unsafe.Pointer(&info.cursorPosition))))

	return nil
}

// windowsGetTermSize returns the dimensions of the given terminal.
// the code is taken from "golang.org/x/crypto/ssh/terminal"
func windowsGetTermSize() (width, height int, err error) {
	var info consoleScreenBufferInfo
	_, _, e := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(syscall.Stdout), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return 0, 0, error(e)
	}
	return int(info.size.x), int(info.size.y), nil
}
