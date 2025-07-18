package manifest

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ManifestEntry represents a single line in the manifest file.
type ManifestEntry struct {
	StartTime float64
	EndTime   float64
	Speaker   string
	FilePath  string
}

// Parse reads and parses the manifest file at the given path.
func Parse(path string) ([]ManifestEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	// It captures start/end times (integer or float), speaker, and file path.
	// It accepts both hyphen (-) and en dash (–) as separators.
	re := regexp.MustCompile(`^\[(\d+(?:\.\d+)?)s[–-](\d+(?:\.\d+)?)s\]\s+\((.+)\)\s+(.+)$`)

	var entries []ManifestEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) != 5 {
			return nil, fmt.Errorf("failed to parse line: %q", line)
		}

		startTime, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}

		endTime, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}

		entry := ManifestEntry{
			StartTime: startTime,
			EndTime:   endTime,
			Speaker:   matches[3],
			FilePath:  matches[4],
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	return entries, nil
}

// Write writes a slice of ManifestEntry structs to a file at the given path.
func Write(path string, entries []ManifestEntry) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range entries {
		// Using ".1f" for consistency with the example format.
		line := fmt.Sprintf("[%.1fs–%.1fs] (%s) %s\n", entry.StartTime, entry.EndTime, entry.Speaker, entry.FilePath)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write to manifest: %w", err)
		}
	}

	return writer.Flush()
}
