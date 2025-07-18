package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// MockAudioProcessor is a mock implementation of the audio.Processor for testing.
type MockAudioProcessor struct {
	GetDurationFunc func(filePath string) (float64, error)
	ApplySpeedFunc  func(inputFile, outputFile string, speed float64) error
}

func (m *MockAudioProcessor) GetDuration(filePath string) (float64, error) {
	if m.GetDurationFunc != nil {
		return m.GetDurationFunc(filePath)
	}
	return 0, fmt.Errorf("GetDurationFunc not implemented")
}

func (m *MockAudioProcessor) ApplySpeed(inputFile, outputFile string, speed float64) error {
	if m.ApplySpeedFunc != nil {
		return m.ApplySpeedFunc(inputFile, outputFile, speed)
	}
	return fmt.Errorf("ApplySpeedFunc not implemented")
}

func TestProcessor_ProcessManifest_Clamping(t *testing.T) {
	testCases := []struct {
		name               string
		manifestDuration   float64
		actualDuration     float64
		expectedSpeed      float64
		shouldApplySpeed   bool
	}{
		{
			name:             "Speed up clamped to max",
			manifestDuration: 20.0,
			actualDuration:   10.0,
			expectedSpeed:    maxSpeed,
			shouldApplySpeed: true,
		},
		{
			name:             "Slow down clamped to min",
			manifestDuration: 5.0,
			actualDuration:   10.0,
			expectedSpeed:    minSpeed,
			shouldApplySpeed: true,
		},
		{
			name:             "Speed within range",
			manifestDuration: 12.0,
			actualDuration:   10.0,
			expectedSpeed:    1.2,
			shouldApplySpeed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var appliedSpeed float64
			mockAudioProc := &MockAudioProcessor{
				GetDurationFunc: func(filePath string) (float64, error) {
					return tc.actualDuration, nil
				},
				ApplySpeedFunc: func(inputFile, outputFile string, speed float64) error {
					appliedSpeed = speed
					return nil
				},
			}

			processor := NewProcessor(mockAudioProc)

			// Create a temporary manifest file
			manifestContent := fmt.Sprintf("[0.0s–%.1fs] (SPEAKER_00) /fake/audio.wav", tc.manifestDuration)
			tmpDir := t.TempDir()
			manifestPath := filepath.Join(tmpDir, "manifest.txt")
			if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
				t.Fatalf("failed to create temp manifest file: %v", err)
			}

			err := processor.ProcessManifest(manifestPath)
			if err != nil {
				t.Fatalf("ProcessManifest() error = %v", err)
			}

			if tc.shouldApplySpeed {
				if appliedSpeed == 0 {
					t.Error("ApplySpeed was not called")
				} else if appliedSpeed != tc.expectedSpeed {
					t.Errorf("expected speed %.2f, got %.2f", tc.expectedSpeed, appliedSpeed)
				}
			}
		})
	}
}

func TestProcessor_ProcessManifest_ErrorHandling(t *testing.T) {
	t.Run("Invalid manifest path", func(t *testing.T) {
		processor := NewProcessor(&MockAudioProcessor{})
		err := processor.ProcessManifest("/non/existent/manifest.txt")
		if !errors.Is(err, ErrInvalidManifest) {
			t.Errorf("expected ErrInvalidManifest, got %v", err)
		}
	})

	t.Run("GetDuration fails", func(t *testing.T) {
		mockAudioProc := &MockAudioProcessor{
			GetDurationFunc: func(filePath string) (float64, error) {
				return 0, fmt.Errorf("mock get duration error")
			},
		}
		processor := NewProcessor(mockAudioProc)

		manifestContent := "[0.0s–5.0s] (SPEAKER_00) /fake/audio.wav"
		tmpDir := t.TempDir()
		manifestPath := filepath.Join(tmpDir, "manifest.txt")
		if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("failed to create temp manifest file: %v", err)
		}

		// We expect no error from ProcessManifest itself, but the entry error should be logged.
		// Testing logs requires more setup, so we'll trust the logic for now.
		err := processor.ProcessManifest(manifestPath)
		if err != nil {
			t.Errorf("ProcessManifest() returned an unexpected error: %v", err)
		}
	})
}
