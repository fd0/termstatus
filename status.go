package termstatus

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
)

type Term struct {
	dst    Terminal
	buf    *bytes.Buffer
	msg    chan message
	status chan message
}

type Terminal interface {
	io.Writer
	Fd() uintptr
}

type message struct {
	buf []byte
	ch  chan<- response
}

type response struct {
	n   int
	err error
}

func New(ctx context.Context, dst Terminal) *Term {
	t := &Term{
		buf:    bytes.NewBuffer(nil),
		dst:    dst,
		msg:    make(chan message),
		status: make(chan message),
	}

	go t.run(ctx)

	return t
}

func countLines(buf []byte) int {
	lines := 0
	sc := bufio.NewScanner(bytes.NewReader(buf))
	for sc.Scan() {
		lines++
	}
	return lines
}

func (t *Term) run(ctx context.Context) {
	statusBuf := bytes.NewBuffer(nil)
	statusLines := 0
	for {
		select {
		case <-ctx.Done():
			t.undoStatus(statusLines)
			return
		case msg := <-t.msg:
			t.undoStatus(statusLines)

			n, err := t.dst.Write(msg.buf)
			if err != nil {
				msg.ch <- response{n: n, err: err}
				continue
			}

			_, err = t.dst.Write(statusBuf.Bytes())
			if err != nil {
				msg.ch <- response{n: n, err: err}
				continue
			}

			msg.ch <- response{n: n}

		case msg := <-t.status:
			buf := bytes.TrimRight(msg.buf, "\n")
			buf = append(buf, '\r')
			t.undoStatus(statusLines)

			lines := countLines(buf)

			_, err := t.dst.Write(buf)
			if err != nil {
				msg.ch <- response{err: err}
				continue
			}

			statusBuf.Reset()
			statusBuf.Write(buf)

			statusLines = lines

			msg.ch <- response{}
		}
	}
}

const (
	moveCursorHome = "\r"
	moveCursorUp   = "\x1b[1A"
	clearLine      = "\x1b[2K"
)

func (t *Term) undoStatus(lines int) error {
	if lines == 0 {
		return nil
	}

	// clear current line
	_, err := t.dst.Write([]byte(moveCursorHome + clearLine))
	if err != nil {
		return err
	}

	lines--

	const clearLine = moveCursorHome + moveCursorUp + clearLine

	for ; lines > 0; lines-- {
		// clear current line and move on line up
		_, err := t.dst.Write([]byte(clearLine))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Term) updateStatus() {
	buf := t.buf.Bytes()

	if len(buf) == 0 {
		return
	}

	fmt.Fprintf(t.dst, "\x1b[2K\n")

	if buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}
	buf = append(buf, '\r')

	lines := 0
	sc := bufio.NewScanner(bytes.NewReader(buf))
	for sc.Scan() {
		lines++
	}

	t.dst.Write(buf)

	fmt.Fprintf(t.dst, "\x1b[%dA", lines)
}

func (t *Term) Write(p []byte) (int, error) {
	ch := make(chan response, 1)
	t.msg <- message{buf: p, ch: ch}
	res := <-ch
	return res.n, res.err
}

func (t *Term) SetStatus(p []byte) error {
	ch := make(chan response, 1)
	t.status <- message{buf: p, ch: ch}
	res := <-ch
	return res.err
}
