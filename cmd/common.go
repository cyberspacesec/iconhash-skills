package cmd

import (
	"time"
)

// Version information
var (
	Version   = "dev"
	BuildDate = "unknown"
	BuildHash = "unknown"
)

// Global flags
var (
	Debug        bool
	Uint32Flag   bool
	URL          string
	FilePath     string
	Base64Path   string
	UserAgent    string
	Engine       string // Search engine format: plain, fofa, shodan, censys, quake, zoomeye, hunter
	SkipVerify   bool
	Timeout      time.Duration
	OutputFormat string
)

// Server flags
var (
	Host         string
	Port         int
	AuthToken    string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	ServerProxy  string // Proxy for server mode
)

// Batch/Output flags
var (
	Proxy         string
	OutputFile    string
	InputFile     string
	FingerprintDB string // Path to custom fingerprint JSON database
)

// Fingerprints update flags
var (
	fpUpdateOutput  string
	fpUpdateURL     string
	fpUpdateReplace bool
)
