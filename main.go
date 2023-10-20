package main

import (
	"fmt"
	"github.com/bariiss/transfersh/lib"

	c "github.com/bariiss/transfersh/lib/config"
	ct "github.com/bariiss/transfersh/lib/content"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

// Path: main.go

var version = "0.1.5"

var rootCmd = &cobra.Command{
	Use:   "transfersh [file|directory]",
	Short: "Transfersh files or directories.",
	Long: `transfersh is a CLI tool for uploading files or directories.
Given a file or directory path, it will upload the content and 
provide a URL for download. If a directory path is provided,
it will be compressed as a .zip file and then uploaded.`,
	Version: version,
	Args:    cobra.ExactArgs(1),
	Run:     executeTransfer,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeTransfer(cmd *cobra.Command, args []string) {
	loadConfig, err := c.LoadConfig()
	if err != nil {
		fmt.Println("Error loading loadConfig:", err)
		return
	}

	file := args[0]                                  // file or directory path
	if _, err := os.Stat(file); os.IsNotExist(err) { // check if file or directory exists
		fmt.Printf("%s: No such file or directory\n", file)
		return
	}

	fileName := filepath.Base(file)              // get file name
	reader, size, err := ct.PrepareContent(file) // prepare content
	if err != nil {
		fmt.Println("Error preparing content:", err)
		return
	}

	if info, err := os.Stat(file); err == nil && info.IsDir() { // check if file is directory
		fileName += ".zip" // add .zip extension
	}

	lib.Upload(reader, fileName, loadConfig, size) // upload and print
}
