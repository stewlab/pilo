package spinner

import (
	"fmt"
	"time"
)

// Spinner represents the loading spinner.
type Spinner struct {
	stopChan  chan struct{}
	startTime time.Time
	message   string
}

// NewSpinner creates and starts a new spinner.
func NewSpinner(message string) *Spinner {
	s := &Spinner{
		stopChan:  make(chan struct{}),
		startTime: time.Now(),
		message:   message,
	}

	return s
}

// Start starts the spinner.
func (s *Spinner) Start() {
	go func() {
		for {
			for _, r := range `-\|/` {
				select {
				case <-s.stopChan:
					return
				default:
					fmt.Printf("\r%s %c ", s.message, r)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	duration := time.Since(s.startTime)
	if duration < 500*time.Millisecond {
		time.Sleep(500*time.Millisecond - duration)
	}
	close(s.stopChan)
	fmt.Printf("\r%s Done!\n", s.message)
}
