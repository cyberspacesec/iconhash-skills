package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// SilenceUsageOnError silences usage output on error for a command.
// Use this with RunE commands so that runtime errors don't print usage.
func SilenceUsageOnError(cmd *cobra.Command) {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
}

// wrapError wraps an error with a formatted message for consistent error output.
// Cobra will print the error when RunE returns a non-nil error.
func wrapError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
