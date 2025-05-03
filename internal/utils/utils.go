package utils

import (
	"encoding/binary"
	"strings"
)

// Itob converts an int to a byte slice for BoltDB keys
func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// Btoi converts a byte slice to an int for BoltDB keys
func Btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

// Contains checks if a string contains a substring (case-insensitive)
func Contains(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

// NormalizeTag removes spaces and converts to lowercase
func NormalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

// ContainsTag checks if a tag slice contains a specific tag
func ContainsTag(tags []string, tag string) bool {
	normalizedTag := NormalizeTag(tag)
	for _, t := range tags {
		if NormalizeTag(t) == normalizedTag {
			return true
		}
	}
	return false
}

// RemoveDuplicateTags removes duplicate tags from a slice
func RemoveDuplicateTags(tags []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, tag := range tags {
		normalized := NormalizeTag(tag)
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}
