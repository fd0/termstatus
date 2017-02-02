package main

import (
	"context"
	"fmt"
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
			time.Sleep(300 * time.Millisecond)
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

		err := t.SetStatus(status)
		if err != nil {
			panic(err)
		}
		time.Sleep(50 * time.Millisecond)

	}
}
