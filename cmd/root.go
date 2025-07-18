package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sync-audio",
	Short: "A CLI tool to synchronize audio files based on a manifest.",
	Long: `sync-audio is a command-line tool designed to process and synchronize audio
files. It can adjust audio speed to match timestamps provided in a manifest file.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
