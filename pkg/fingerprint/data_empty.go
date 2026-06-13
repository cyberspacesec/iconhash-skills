//go:build !embed_fingerprints

package fingerprint

// DefaultFingerprints is empty in lightweight (non-embedded) build mode.
// The fingerprint database will be loaded from an external JSON file.
// Use `iconhash fingerprints update` to download the database,
// or set ICONHASH_FINGERPRINT_DB environment variable.
var DefaultFingerprints = []FingerprintEntry{}
