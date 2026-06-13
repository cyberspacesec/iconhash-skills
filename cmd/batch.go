package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type batchResult struct {
	URL  string `json:"url"`
	Hash string `json:"hash,omitempty"`
	Err  string `json:"error,omitempty"`
}

// NewBatchCommand creates the batch command
func NewBatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch calculate hashes from URLs in a file or stdin",
		Long: `Batch calculate favicon hashes from a list of URLs.

Reads URLs from a file (--input) or stdin (pipe), calculates the favicon hash
for each URL, and outputs the results.

Examples:
  iconhash batch -i urls.txt
  iconhash batch -i urls.txt -o results.json
  cat urls.txt | iconhash batch
  iconhash batch -i urls.txt --engine fofa --proxy socks5://127.0.0.1:1080`,
		RunE: runBatch,
	}

	SilenceUsageOnError(cmd)

	return cmd
}

// runBatch handles the batch command execution
func runBatch(cmd *cobra.Command, args []string) error {
	var reader io.Reader

	if InputFile != "" {
		f, err := os.Open(InputFile)
		if err != nil {
			return wrapError("error opening input file: %w", err)
		}
		defer f.Close()
		reader = f
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input provided. Use --input flag or pipe URLs via stdin")
		}
		reader = os.Stdin
	}

	return runBatchFromReader(reader)
}

// runBatchFromReader processes URLs from a reader and outputs hash results
func runBatchFromReader(r io.Reader) error {
	options := buildHashOptions()
	h := hasher.New(options)

	format := getEngineFormat()

	// Use the SDK's BatchHashFromReader method
	results, err := h.BatchHashFromReader(context.Background(), r, 0) // sequential by default
	if err != nil {
		return wrapError("error reading input: %w", err)
	}

	if OutputFile != "" {
		// Convert SDK results to output format
		var outputResults []batchResult
		for _, r := range results {
			br := batchResult{URL: r.URL}
			if r.Err != nil {
				br.Err = r.Err.Error()
			} else {
				br.Hash = util.FormatHash(r.Hash, format)
			}
			outputResults = append(outputResults, br)
		}
		writeBatchOutput(outputResults)
	} else {
		for _, r := range results {
			if r.Err != nil {
				fmt.Fprintf(os.Stderr, "❌ %s: %v\n", r.URL, r.Err)
			} else {
				fmt.Printf("%s %s\n", r.URL, util.FormatHash(r.Hash, format))
			}
		}
	}

	return nil
}

func buildHashOptions() *hasher.HashOptions {
	if Proxy != "" {
		opts, err := hasher.NewOptionsWithProxy(Proxy, Timeout, SkipVerify)
		if err == nil {
			opts.UseUint32 = Uint32Flag
			opts.UserAgent = UserAgent
			return opts
		}
		// Fall through to default if proxy parsing fails
	}
	return &hasher.HashOptions{
		UseUint32:          Uint32Flag,
		RequestTimeout:     Timeout,
		InsecureSkipVerify: SkipVerify,
		UserAgent:          UserAgent,
	}
}

func writeBatchOutput(results []batchResult) {
	if strings.HasSuffix(OutputFile, ".csv") {
		f, err := os.Create(OutputFile)
		if err != nil {
			color.Red("❌ Error creating output file: %v", err)
			return
		}
		defer f.Close()
		f.WriteString("url,hash,error\n")
		for _, r := range results {
			fmt.Fprintf(f, "%q,%q,%q\n", r.URL, r.Hash, r.Err)
		}
	} else {
		data, _ := json.MarshalIndent(results, "", "  ")
		os.WriteFile(OutputFile, data, 0644)
	}
}
