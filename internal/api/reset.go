package api

import (
	"pilo/internal/config"
)

// Reset clears the Pilo configuration and logs.
func Reset() error {
	// Clear the logs
	config.SetLogs([]string{})

	return nil
}
