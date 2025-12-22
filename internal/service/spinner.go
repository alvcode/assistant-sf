package service

import (
	"fmt"
	"os"
	"time"
)

func StartSpinner(msg string) func() {
	done := make(chan struct{})
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				_, err := fmt.Fprintf(os.Stdout, "\r\033[K%s %s", spinner[i%len(spinner)], msg)
				if err != nil {
					return
				}
				i++
				time.Sleep(120 * time.Millisecond)
			}
		}
	}()

	return func() {
		close(done)
		_, err := fmt.Fprint(os.Stdout, "\r\033[K")
		if err != nil {
			return
		}
	}
}
