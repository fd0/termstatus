package termstatus

import (
	"bufio"
	"bytes"
	"context"
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

// run listens on the channels and updates the terminal screen.
func (t *Term) run(ctx context.Context) {
	statusBuf := bytes.NewBuffer(nil)
	statusLines := 0
	for {
		select {
		case <-ctx.Done():
			t.undoStatus(statusLines)
			return
		case msg := <-t.msg:
			err := t.undoStatus(statusLines)
			if err != nil {
				msg.ch <- response{err: err}
				continue
			}

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
			err := t.undoStatus(statusLines)
			if err != nil {
				msg.ch <- response{err: err}
				continue
			}

			buf := bytes.TrimRight(msg.buf, "\n")
			lines := countLines(buf)

			_, err = t.dst.Write(buf)
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

func (t *Term) undoStatus(lines int) error {
	if lines == 0 {
		return nil
	}

	lines--
	return clearLines(t.dst, lines)
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
