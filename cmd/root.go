package cmd

import (
	"fmt"
	"time"

	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command
var RootCmd = &cobra.Command{
	Use:     "iconhash [command]",
	Short:   "Icon Hash Calculator - A tool for cybersecurity reconnaissance",
	Version: Version,
	Long: `Icon Hash Calculator - A tool for cybersecurity reconnaissance

Calculate the MMH3 hash of a favicon.ico file for use with search engines like
Fofa, Shodan, Censys, Quake, ZoomEye, and Hunter. This tool can process
favicons from URLs, files, or base64 data.

Examples:
  iconhash url https://example.com/favicon.ico        # Hash from URL
  iconhash file favicon.ico                            # Hash from file
  iconhash base64 base64file.txt                       # Hash from base64 file
  iconhash url https://example.com --engine fofa       # Hash with Fofa format
  iconhash identify https://example.com                # Discover + identify
  iconhash lookup -- -305179312                        # Lookup fingerprint
  iconhash batch -i urls.txt                           # Batch process URLs
  iconhash server -p 8080                              # Start API server`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// PrintLogo prints the ASCII art logo
func PrintLogo() {
	cyan := color.New(color.FgCyan).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	logo := `
  ██▓ ▄████▄   ▒█████   ███▄    █     ██░ ██  ▄▄▄       ██████  ██░ ██
 ▓██▒▒██▀ ▀█  ▒██▒  ██▒ ██ ▀█   █    ▓██░ ██▒▒████▄   ▒██    ▒ ▓██░ ██▒
 ▒██▒▒▓█    ▄ ▒██░  ██▒▓██  ▀█ ██▒   ▒██▀▀██░▒██  ▀█▄ ░ ▓██▄   ▒██▀▀██░
 ░██░▒▓▓▄ ▄██▒▒██   ██░▓██▒  ▐▌██▒   ░▓█ ░██ ░██▄▄▄▄██  ▒   ██▒░▓█ ░██
 ░██░▒ ▓███▀ ░░ ████▓▒░▒██░   ▓██░   ░▓█▒░██▓ ▓█   ▓██▒▒██████▒▒░▓█▒░██▓
 ░▓  ░ ░▒ ▒  ░░ ▒░▒░▒░ ░ ▒░   ▒ ▒     ▒ ░░▒░▒ ▒▒   ▓▒█░▒ ▒▓▒ ▒ ░ ▒ ░░▒░▒
  ▒ ░  ░  ▒     ░ ▒ ▒░ ░ ░░   ░ ▒░    ▒ ░▒░ ░  ▒   ▒▒ ░░ ░▒  ░ ░ ▒ ░▒░ ░
  ▒ ░░        ░ ░ ░ ▒     ░   ░ ░     ░  ░░ ░  ░   ▒   ░  ░  ░   ░  ░░ ░
  ░  ░ ░          ░ ░           ░     ░  ░  ░      ░  ░      ░   ░  ░  ░
     ░
`
	coloredLogo := cyan(logo)
	fmt.Println(coloredLogo)
	fmt.Printf("%s %s - Version %s\n",
		blue("IconHash Calculator"),
		cyan("by Cyberspace Security"),
		blue(Version))
	fmt.Printf("Build Date: %s | Hash: %s\n\n", BuildDate, BuildHash)
}

// getEngineFormat resolves the --engine flag to a util.OutputFormat.
// Supports: plain, fofa, shodan, censys, quake, zoomeye, hunter.
// Defaults to FormatPlain if empty or unrecognized.
func getEngineFormat() util.OutputFormat {
	switch Engine {
	case "fofa":
		return util.FormatFofa
	case "shodan":
		return util.FormatShodan
	case "censys":
		return util.FormatCensys
	case "quake":
		return util.FormatQuake
	case "zoomeye":
		return util.FormatZoomEye
	case "hunter":
		return util.FormatHunter
	case "plain":
		return util.FormatPlain
	default:
		return util.FormatPlain
	}
}

// Initialize function to set up all commands and flags
func Initialize() {
	// Add all subcommands
	RootCmd.AddCommand(NewURLCommand())
	RootCmd.AddCommand(NewFileCommand())
	RootCmd.AddCommand(NewBase64Command())
	RootCmd.AddCommand(NewServerCommand())
	RootCmd.AddCommand(NewDiscoverCommand())
	RootCmd.AddCommand(NewLookupCommand())
	RootCmd.AddCommand(NewFingerprintsCommand())
	RootCmd.AddCommand(NewIdentifyCommand())
	RootCmd.AddCommand(NewBatchCommand())

	// Define global persistent flags (available to all subcommands)
	RootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "Enable debug output")
	RootCmd.PersistentFlags().BoolVarP(&Uint32Flag, "uint32", "n", false, "Output hash as uint32 instead of int32")
	RootCmd.PersistentFlags().StringVarP(&UserAgent, "user-agent", "a", "", "User agent for HTTP requests")
	RootCmd.PersistentFlags().StringVarP(&Engine, "engine", "e", "", "Search engine format (plain, fofa, shodan, censys, quake, zoomeye, hunter)")
	RootCmd.PersistentFlags().BoolVarP(&SkipVerify, "insecure", "k", false, "Skip TLS certificate verification")
	RootCmd.PersistentFlags().DurationVarP(&Timeout, "timeout", "t", 30*time.Second, "HTTP request timeout")
	RootCmd.PersistentFlags().StringVarP(&OutputFormat, "format", "", "text", "Output format (text, json, csv)")
	RootCmd.PersistentFlags().StringVar(&Proxy, "proxy", "", "HTTP/SOCKS5 proxy URL (e.g. socks5://127.0.0.1:1080)")
	RootCmd.PersistentFlags().StringVarP(&OutputFile, "output", "o", "", "Output file path (supports .json and .csv)")
	RootCmd.PersistentFlags().StringVarP(&InputFile, "input", "i", "", "Input file with URLs (one per line)")
	RootCmd.PersistentFlags().StringVar(&FingerprintDB, "fingerprint-db", "", "Path to custom fingerprint JSON database")

	// URL/file/b64 flags are now subcommand-specific, but we keep them as persistent
	// for backward compatibility with the shorthand pattern: iconhash -u <url>
	RootCmd.PersistentFlags().StringVarP(&URL, "url", "u", "", "URL to favicon")
	RootCmd.PersistentFlags().StringVarP(&FilePath, "file", "f", "", "Path to favicon file")
	RootCmd.PersistentFlags().StringVarP(&Base64Path, "b64", "b", "", "Path to file containing base64 encoded favicon")
}
