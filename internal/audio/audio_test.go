package audio

import (
	"fmt"
	"testing"
)

// MockAudioProcessor is a mock implementation of the AudioProcessor interface for testing.
type MockAudioProcessor struct {
	GetDurationFunc  func(filePath string) (float64, error)
	ApplySpeedFunc   func(inputFile, outputFile string, speed float64) error
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

func TestFFmpegProcessor_ApplySpeed_UnsupportedSpeed(t *testing.T) {
	processor := NewFFmpegProcessor()

	// Test speed factor less than 0.5
	err := processor.ApplySpeed("input.wav", "output.wav", 0.4)
	if err == nil {
		t.Error("expected an error for a speed factor less than 0.5, but got nil")
	}

	// Test speed factor greater than 2.0
	err = processor.ApplySpeed("input.wav", "output.wav", 2.1)
	if err == nil {
		t.Error("expected an error for a speed factor greater than 2.0, but got nil")
	}
}
