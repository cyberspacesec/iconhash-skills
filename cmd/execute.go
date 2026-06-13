package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once.
func Execute() {
	// Initialize all commands and flags
	Initialize()

	// Print logo (before any Cobra output)
	PrintLogo()

	// Execute the root command
	err := RootCmd.Execute()
	if err != nil {
		color.Red("❌ %v", err)
		os.Exit(1)
	}
}

// fatalf prints a formatted error message and exits.
// Use this ONLY in Execute() or main(), never in RunE functions.
func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
