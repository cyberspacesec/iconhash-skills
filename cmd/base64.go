package cmd

import (
	"fmt"
	"os"

	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewBase64Command creates the base64 command
func NewBase64Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base64 [filepath]",
		Short: "Generate hash from a base64 encoded file",
		Long: `Generate a favicon hash from a base64 encoded file.

This command will read the specified file containing base64 encoded data and calculate its hash.
The hash can be formatted for use with search engines like Fofa, Shodan, Censys,
Quake, ZoomEye, and Hunter.

Examples:
  iconhash base64 encoded.txt
  iconhash base64 -b /path/to/encoded.txt --engine shodan
  iconhash base64 icon_b64.txt --uint32 --engine fofa`,
		RunE: runBase64,
		Args: func(cmd *cobra.Command, args []string) error {
			// If filepath is provided as positional arg, set it in the flags
			if len(args) > 0 {
				// If base64 path already set via flag, don't override
				if Base64Path == "" {
					Base64Path = args[0]
				}
			}

			// Validate we have a filepath
			if Base64Path == "" {
				return fmt.Errorf("base64 filepath is required. Provide it as an argument or with --b64 flag")
			}

			// Check if file exists
			_, err := os.Stat(Base64Path)
			if err != nil {
				return fmt.Errorf("file not found: %s", Base64Path)
			}

			return nil
		},
	}

	SilenceUsageOnError(cmd)

	return cmd
}

// runBase64 handles the base64 command execution
func runBase64(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Base64 File: %s\n", Base64Path)
		fmt.Fprintf(cmd.ErrOrStderr(), "🔧 Options: uint32=%v\n", options.UseUint32)
	}

	// Read file data
	fmt.Fprintf(cmd.ErrOrStderr(), "📂 Reading base64 file %s...\n", Base64Path)
	fileData, err := os.ReadFile(Base64Path)
	if err != nil {
		return wrapError("error reading file: %w", err)
	}

	// Calculate hash
	fmt.Fprintf(cmd.ErrOrStderr(), "🧮 Calculating hash...\n")
	hash, err := h.HashFromBase64(string(fileData))
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
