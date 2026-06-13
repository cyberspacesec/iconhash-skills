package cmd

import (
	"fmt"
	"os"

	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewFileCommand creates the file command
func NewFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file [filepath]",
		Short: "Generate hash from a file",
		Long: `Generate a favicon hash from a local file.

This command will read the specified favicon file and calculate its hash.
The hash can be formatted for use with search engines like Fofa, Shodan, Censys,
Quake, ZoomEye, and Hunter.

Examples:
  iconhash file favicon.ico
  iconhash file -f /path/to/favicon.ico --engine shodan
  iconhash file icon.png --uint32 --engine fofa`,
		RunE: runFile,
		Args: func(cmd *cobra.Command, args []string) error {
			// If filepath is provided as positional arg, set it in the flags
			if len(args) > 0 {
				// If filepath already set via flag, don't override
				if FilePath == "" {
					FilePath = args[0]
				}
			}

			// Validate we have a filepath
			if FilePath == "" {
				return fmt.Errorf("filepath is required. Provide it as an argument or with --file flag")
			}

			// Check if file exists
			_, err := os.Stat(FilePath)
			if err != nil {
				return fmt.Errorf("file not found: %s", FilePath)
			}

			return nil
		},
	}

	SilenceUsageOnError(cmd)

	return cmd
}

// runFile handles the file command execution
func runFile(cmd *cobra.Command, args []string) error {
	// Create a new hasher with the options from root command
	options := &hasher.HashOptions{
		UseUint32:          Uint32Flag,
		RequestTimeout:     Timeout,
		InsecureSkipVerify: SkipVerify,
		UserAgent:          UserAgent,
	}
	h := hasher.New(options)

	// Debug info if enabled
	if Debug {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 File: %s\n", FilePath)
		fmt.Fprintf(cmd.ErrOrStderr(), "🔧 Options: uint32=%v\n", options.UseUint32)
	}

	// Read file data
	fmt.Fprintf(cmd.ErrOrStderr(), "📂 Reading file %s...\n", FilePath)
	fileData, err := os.ReadFile(FilePath)
	if err != nil {
		return wrapError("error reading file: %w", err)
	}

	// Calculate hash
	fmt.Fprintf(cmd.ErrOrStderr(), "🧮 Calculating hash...\n")
	hash, err := h.HashFromBytes(fileData)
	if err != nil {
		return wrapError("error calculating hash: %w", err)
	}

	// Get format from --engine flag
	format := getEngineFormat()

	// Format the hash
	formatted := util.FormatHash(hash, format)

	// Print hash with color
	boldGreen := color.New(color.FgGreen, color.Bold)
	boldGreen.Println("✅ Hash calculated successfully!")
	fmt.Println()

	boldCyan := color.New(color.FgCyan, color.Bold)
	boldCyan.Printf("Hash: ")
	fmt.Println(hash)

	if format != util.FormatPlain {
		boldCyan.Printf("Formatted: ")
		fmt.Println(formatted)
	}

	return nil
}
