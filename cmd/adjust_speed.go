package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/viniciusrtf/sync-audio-with-timestamps/internal/audio"
	"github.com/viniciusrtf/sync-audio-with-timestamps/pkg/core"
)

var manifestPath string

var adjustSpeedCmd = &cobra.Command{
	Use:   "adjust-speed",
	Short: "Adjusts the speed of audio files based on a manifest.",
	Long: `This command reads a manifest file containing timestamps and audio file paths.
It calculates the necessary speed adjustment for each audio file to match the
target duration and applies it using ffmpeg.`,
	Run: func(cmd *cobra.Command, args []string) {
		if manifestPath == "" {
			log.Fatal("manifest path is required")
		}

		audioProcessor := audio.NewFFmpegProcessor()
		coreProcessor := core.NewProcessor(audioProcessor)

		if err := coreProcessor.ProcessManifest(manifestPath); err != nil {
			// The core processor logs errors for individual entries, so we only need to handle fatal errors.
			// A fatal error can be an invalid manifest or a failure to write the new synced manifest.
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(adjustSpeedCmd)
	adjustSpeedCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "Path to the manifest file (required)")
	adjustSpeedCmd.MarkFlagRequired("manifest")
}
