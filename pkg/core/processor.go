package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/viniciusrtf/sync-audio-with-timestamps/internal/audio"
	"github.com/viniciusrtf/sync-audio-with-timestamps/internal/manifest"
)

const (
	minSpeed = 0.8
	maxSpeed = 1.5
	// Epsilon is a small value to handle floating-point inaccuracies.
	epsilon = 0.01
)

// Processor handles the core logic of processing the manifest entries.
type Processor struct {
	audioProc audio.Processor
}

// NewProcessor creates a new core Processor.
func NewProcessor(audioProc audio.Processor) *Processor {
	return &Processor{
		audioProc: audioProc,
	}
}

// ProcessManifest processes the manifest file, adjusts audio speed, and writes a new manifest for the synced files.
func (p *Processor) ProcessManifest(manifestPath string) error {
	entries, err := manifest.Parse(manifestPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidManifest, err)
	}

	var syncedEntries []manifest.ManifestEntry
	for _, entry := range entries {
		newEntry, err := p.processEntry(entry)
		if err != nil {
			// Log the error and continue to the next entry, so one failure doesn't stop the whole process.
			log.Printf("Skipping entry for %s: %v", entry.FilePath, err)
		} else {
			syncedEntries = append(syncedEntries, newEntry)
		}
	}

	// If any entries were successfully processed, write the new manifest.
	if len(syncedEntries) > 0 {
		syncedManifestPath := getSyncedManifestPath(manifestPath)
		if err := manifest.Write(syncedManifestPath, syncedEntries); err != nil {
			return fmt.Errorf("failed to write synced manifest: %w", err)
		}
		fmt.Printf("\nSuccessfully created synced manifest: %s\n", syncedManifestPath)
	} else {
		fmt.Println("\nNo audio files were successfully processed; synced manifest not created.")
	}

	return nil
}

// processEntry handles the logic for a single manifest entry.
// It returns a new ManifestEntry with the updated file path on success.
func (p *Processor) processEntry(entry manifest.ManifestEntry) (manifest.ManifestEntry, error) {
	fmt.Printf("Processing %s...\n", entry.FilePath)

	manifestDuration := entry.EndTime - entry.StartTime
	if manifestDuration <= 0 {
		return manifest.ManifestEntry{}, fmt.Errorf("invalid duration in manifest (%.2fs)", manifestDuration)
	}

	actualDuration, err := p.audioProc.GetDuration(entry.FilePath)
	if err != nil {
		return manifest.ManifestEntry{}, fmt.Errorf("%w: %w", ErrProcessingEntry, err)
	}

	if actualDuration == 0 {
		return manifest.ManifestEntry{}, fmt.Errorf("actual duration is zero")
	}

	speedFactor := actualDuration / manifestDuration
	fmt.Printf("  Manifest duration: %.2fs\n", manifestDuration)
	fmt.Printf("  Actual duration:   %.2fs\n", actualDuration)
	fmt.Printf("  Original speed factor: %.2f\n", speedFactor)

	// Clamp the speed factor to the allowed range.
	clampedSpeed := clamp(speedFactor, minSpeed, maxSpeed)
	if clampedSpeed != speedFactor {
		fmt.Printf("  Clamped speed factor:  %.2f\n", clampedSpeed)
	}

	outputFilePath := getOutputFilePath(entry.FilePath)
	if err := p.audioProc.ApplySpeed(entry.FilePath, outputFilePath, clampedSpeed); err != nil {
		return manifest.ManifestEntry{}, fmt.Errorf("%w: %w", ErrProcessingEntry, err)
	}

	fmt.Printf("  Successfully created %s\n", outputFilePath)

	// Return a new entry pointing to the synced file.
	return manifest.ManifestEntry{
		StartTime: entry.StartTime,
		EndTime:   entry.EndTime,
		Speaker:   entry.Speaker,
		FilePath:  outputFilePath,
	}, nil
}

// BuildFromManifest creates a single audio file by concatenating clips from a manifest.
func (p *Processor) BuildFromManifest(manifestPath, outputPath string) error {
	entries, err := manifest.Parse(manifestPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidManifest, err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("cannot build from an empty manifest")
	}

	// Create a temporary directory for intermediate files.
	tempDir, err := os.MkdirTemp("", "sync-audio-build-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var currentTime float64
	var currentBuildFile string

	for i, entry := range entries {
		fmt.Printf("Step %d/%d: Processing %s\n", i+1, len(entries), entry.FilePath)

		// 1. Calculate and generate silence if there's a gap.
		gap := entry.StartTime - currentTime
		filesToConcat := []string{}

		if i > 0 {
			filesToConcat = append(filesToConcat, currentBuildFile)
		}

		if gap > epsilon {
			fmt.Printf("  Adding %.2fs of silence.\n", gap)
			silenceFile := filepath.Join(tempDir, fmt.Sprintf("silence_%d.wav", i))
			if err := p.audioProc.GenerateSilence(gap, silenceFile); err != nil {
				return fmt.Errorf("failed to generate silence for entry %d: %w", i, err)
			}
			filesToConcat = append(filesToConcat, silenceFile)
		}

		// 2. Add the actual audio clip to the list.
		filesToConcat = append(filesToConcat, entry.FilePath)

		// 3. Concatenate the files.
		nextBuildFile := filepath.Join(tempDir, fmt.Sprintf("build_step_%d.wav", i))
		fmt.Printf("  Concatenating %d files...\n", len(filesToConcat))
		if err := p.audioProc.Concatenate(filesToConcat, nextBuildFile); err != nil {
			return fmt.Errorf("failed to concatenate files for entry %d: %w", i, err)
		}

		// The output of this step becomes the input for the next.
		currentBuildFile = nextBuildFile
		// Update the timeline.
		newDuration, err := p.audioProc.GetDuration(currentBuildFile)
		if err != nil {
			return fmt.Errorf("failed to get duration of intermediate build file: %w", err)
		}
		currentTime = newDuration
		fmt.Printf("  Intermediate track duration: %.2fs\n", currentTime)
	}

	// Finally, move the last intermediate file to the final output path.
	fmt.Printf("\nBuild complete. Moving final file to %s\n", outputPath)
	if err := os.Rename(currentBuildFile, outputPath); err != nil {
		return fmt.Errorf("failed to move final build file: %w", err)
	}

	return nil
}

// getSyncedManifestPath generates the name for the new manifest file.
func getSyncedManifestPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, fmt.Sprintf("%s_synced%s", name, ext))
}

func getOutputFilePath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, fmt.Sprintf("%s_synced%s", name, ext))
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
