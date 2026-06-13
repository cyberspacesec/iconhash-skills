package util

import "fmt"

// OutputFormat represents the format of the hash output
type OutputFormat int

const (
	// FormatPlain outputs the hash as is
	FormatPlain OutputFormat = iota
	// FormatFofa outputs the hash in Fofa search format
	FormatFofa
	// FormatShodan outputs the hash in Shodan search format
	FormatShodan
	// FormatCensys outputs the hash in Censys search format
	FormatCensys
	// FormatQuake outputs the hash in 360 Quake search format
	FormatQuake
	// FormatZoomEye outputs the hash in ZoomEye search format
	FormatZoomEye
	// FormatHunter outputs the hash in Hunter search format
	FormatHunter
)

// ParseFormat parses a format name string into an OutputFormat.
// Supported names: "plain", "fofa", "shodan", "censys", "quake", "zoomeye", "hunter".
// Returns FormatPlain for unrecognized or empty strings.
func ParseFormat(name string) OutputFormat {
	switch name {
	case "fofa":
		return FormatFofa
	case "shodan":
		return FormatShodan
	case "censys":
		return FormatCensys
	case "quake":
		return FormatQuake
	case "zoomeye":
		return FormatZoomEye
	case "hunter":
		return FormatHunter
	case "plain":
		return FormatPlain
	default:
		return FormatPlain
	}
}

// FormatName returns the canonical lowercase name for an OutputFormat.
func FormatName(format OutputFormat) string {
	switch format {
	case FormatPlain:
		return "plain"
	case FormatFofa:
		return "fofa"
	case FormatShodan:
		return "shodan"
	case FormatCensys:
		return "censys"
	case FormatQuake:
		return "quake"
	case FormatZoomEye:
		return "zoomeye"
	case FormatHunter:
		return "hunter"
	default:
		return "unknown"
	}
}

// FormatHash formats a hash value according to the specified output format
func FormatHash(hash string, format OutputFormat) string {
	switch format {
	case FormatFofa:
		return fmt.Sprintf("icon_hash=\"%s\"", hash)
	case FormatShodan:
		return fmt.Sprintf("http.favicon.hash:%s", hash)
	case FormatCensys:
		return fmt.Sprintf("services.http.response.favicons.md5_hash:%s", hash)
	case FormatQuake:
		return fmt.Sprintf("favicon.hash:\"%s\"", hash)
	case FormatZoomEye:
		return fmt.Sprintf("iconhash:\"%s\"", hash)
	case FormatHunter:
		return fmt.Sprintf("web.icon=\"%s\"", hash)
	default:
		return hash
	}
}

// FormatAll returns a map of search engine names to their formatted query strings
// for the given hash. The map keys are: "fofa", "shodan", "censys", "quake", "zoomeye", "hunter".
// This is useful for generating all cyber-space mapping queries at once.
func FormatAll(hash string) map[string]string {
	return map[string]string{
		"fofa":    FormatHash(hash, FormatFofa),
		"shodan":  FormatHash(hash, FormatShodan),
		"censys":  FormatHash(hash, FormatCensys),
		"quake":   FormatHash(hash, FormatQuake),
		"zoomeye": FormatHash(hash, FormatZoomEye),
		"hunter":  FormatHash(hash, FormatHunter),
	}
}

// AllEngines returns the list of all supported search engine names.
func AllEngines() []string {
	return []string{"fofa", "shodan", "censys", "quake", "zoomeye", "hunter"}
}
