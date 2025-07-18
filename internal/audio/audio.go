package audio

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Processor is an interface for audio processing operations.
type Processor interface {
	GetDuration(filePath string) (float64, error)
	ApplySpeed(inputFile, outputFile string, speed float64) error
	GenerateSilence(duration float64, outputFile string) error
	Concatenate(inputFiles []string, outputFile string) error
}

// FFmpegProcessor implements the AudioProcessor interface using ffmpeg.
type FFmpegProcessor struct{}

// NewFFmpegProcessor creates a new FFmpegProcessor.
func NewFFmpegProcessor() *FFmpegProcessor {
	return &FFmpegProcessor{}
}

// GetDuration returns the duration of an audio file in seconds.
func (p *FFmpegProcessor) GetDuration(filePath string) (float64, error) {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 <filePath>
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed with output: %s: %w", string(output), err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// ApplySpeed changes the speed of an audio file and saves it to a new file.
func (p *FFmpegProcessor) ApplySpeed(inputFile, outputFile string, speed float64) error {
	// Defensive check for the atempo filter's supported range.
	// The core processor should handle clamping, but this prevents invalid ffmpeg commands.
	if speed < 0.5 || speed > 2.0 {
		return fmt.Errorf("speed factor %.2f is outside the ffmpeg supported range (0.5-2.0)", speed)
	}

	// ffmpeg -i <inputFile> -filter:a "atempo=<speed>" <outputFile>
	cmd := exec.Command("ffmpeg", "-y", "-i", inputFile, "-filter:a", fmt.Sprintf("atempo=%.2f", speed), outputFile)
	// It's important to capture and wrap the error from ffmpeg if it fails.
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg failed with output: %s: %w", string(output), err)
	}

	return nil
}

// GenerateSilence creates a silent audio file of a given duration.
func (p *FFmpegProcessor) GenerateSilence(duration float64, outputFile string) error {
	// ffmpeg -f lavfi -i anullsrc=r=44100:cl=mono -t <duration> <outputFile>
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", fmt.Sprintf("%.3f", duration), outputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg failed to generate silence: %s: %w", string(output), err)
	}
	return nil
}

// Concatenate joins multiple audio files into a single file.
func (p *FFmpegProcessor) Concatenate(inputFiles []string, outputFile string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided for concatenation")
	}

	var inputs []string
	var filterComplex string
	for i, file := range inputFiles {
		inputs = append(inputs, "-i", file)
		filterComplex += fmt.Sprintf("[%d:a]", i)
	}
	filterComplex += fmt.Sprintf("concat=n=%d:v=0:a=1", len(inputFiles))

	args := []string{"-y"}
	args = append(args, inputs...)
	args = append(args, "-filter_complex", filterComplex, outputFile)

	cmd := exec.Command("ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg failed to concatenate files: %s: %w", string(output), err)
	}
	return nil
}
