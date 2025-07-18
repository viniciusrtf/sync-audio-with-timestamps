package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/viniciusrtf/sync-audio-with-timestamps/internal/audio"
	"github.com/viniciusrtf/sync-audio-with-timestamps/pkg/core"
)

var (
	buildManifestPath string
	buildOutputPath   string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a single audio file from a manifest.",
	Long: `This command takes a manifest of audio clips and concatenates them into a
single audio file. It inserts silence between clips as needed to ensure they
start at the correct timestamps specified in the manifest.`,
	Run: func(cmd *cobra.Command, args []string) {
		audioProcessor := audio.NewFFmpegProcessor()
		coreProcessor := core.NewProcessor(audioProcessor)

		if err := coreProcessor.BuildFromManifest(buildManifestPath, buildOutputPath); err != nil {
			log.Fatalf("Error during build process: %v", err)
		}
		log.Println("Build completed successfully.")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&buildManifestPath, "manifest", "m", "", "Path to the manifest file (required)")
	buildCmd.Flags().StringVarP(&buildOutputPath, "output", "o", "", "Path for the final output audio file (required)")
	buildCmd.MarkFlagRequired("manifest")
	buildCmd.MarkFlagRequired("output")
}
