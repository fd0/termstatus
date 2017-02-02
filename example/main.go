package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fd0/termstatus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	t := termstatus.New(ctx, os.Stdout)

	go func() {
		i := 1
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			fmt.Fprintf(t, "message %v\n", i)
			time.Sleep(800 * time.Millisecond)
			i++
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		status := []byte(fmt.Sprintf("current time: %v\nfoobar line 2\n", time.Now()))
		if rand.Float32() > 0.5 {
			status = append(status, []byte("another line\n")...)
		}
		if rand.Float32() > 0.5 {
			status = append(status, []byte("another line foo\n")...)
		}

		t.SetStatus(status)
		time.Sleep(400 * time.Millisecond)

	}
}
