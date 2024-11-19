package main

import (
	"fmt"
	ct "github.com/bariiss/transfersh/lib/content"
	"github.com/spf13/cobra"
	"os"
)

// Path: main.go

var version = "0.1.8"

var rootCmd = &cobra.Command{
	Use:   "transfersh [file|directory]",
	Short: "Transfersh files or directories.",
	Long: `transfersh is a CLI tool for uploading files or directories.
Given a file or directory path, it will upload the content and 
provide a URL for download. If a directory path is provided,
it will be compressed as a .zip file and then uploaded.`,
	Version: version,
	Args:    cobra.ExactArgs(1),
	Run:     ct.ExecuteTransfer,
}

func init() {
	rootCmd.Flags().StringVar(&ct.MaxDays, "max-days", "", "Maximum number of days before the file is deleted")
	rootCmd.Flags().StringVar(&ct.MaxDownloads, "max-downloads", "", "Maximum number of downloads before the file is deleted")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
