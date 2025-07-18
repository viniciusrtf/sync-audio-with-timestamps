package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	// Create a temporary manifest file for testing
	content := `[0.0s–5.0s] (SPEAKER_00) /path/to/audio/000.wav
[5.7s–8.4s] (SPEAKER_01) /path/to/audio/001.wav
`
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.txt")
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp manifest file: %v", err)
	}

	entries, err := Parse(manifestPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Check the first entry
	if entries[0].StartTime != 0.0 {
		t.Errorf("expected StartTime to be 0.0, got %f", entries[0].StartTime)
	}
	if entries[0].EndTime != 5.0 {
		t.Errorf("expected EndTime to be 5.0, got %f", entries[0].EndTime)
	}
	if entries[0].Speaker != "SPEAKER_00" {
		t.Errorf("expected Speaker to be 'SPEAKER_00', got %s", entries[0].Speaker)
	}
	if entries[0].FilePath != "/path/to/audio/000.wav" {
		t.Errorf("expected FilePath to be '/path/to/audio/000.wav', got %s", entries[0].FilePath)
	}

	// Check the second entry
	if entries[1].StartTime != 5.7 {
		t.Errorf("expected StartTime to be 5.7, got %f", entries[1].StartTime)
	}
	if entries[1].EndTime != 8.4 {
		t.Errorf("expected EndTime to be 8.4, got %f", entries[1].EndTime)
	}
	if entries[1].Speaker != "SPEAKER_01" {
		t.Errorf("expected Speaker to be 'SPEAKER_01', got %s", entries[1].Speaker)
	}
	if entries[1].FilePath != "/path/to/audio/001.wav" {
		t.Errorf("expected FilePath to be '/path/to/audio/001.wav', got %s", entries[1].FilePath)
	}
}

func TestParse_InvalidFile(t *testing.T) {
	_, err := Parse("/non/existent/file")
	if err == nil {
		t.Error("expected an error for a non-existent file, but got nil")
	}
}

func TestParse_InvalidLine(t *testing.T) {
	content := `[0.0s–5.0s] (SPEAKER_00) /path/to/audio/000.wav
invalid line
`
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.txt")
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp manifest file: %v", err)
	}

	_, err := Parse(manifestPath)
	if err == nil {
		t.Error("expected an error for an invalid line, but got nil")
	}
}
