package main

import (
	"context"
	"io"
	"os"

	"github.com/fd0/termstatus"
	"github.com/fd0/termstatus/progress"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t := termstatus.New(ctx, os.Stderr)

	rd := progress.Reader(ctx, os.Stdin, t)

	io.Copy(os.Stdout, rd)
}
