package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

type Config struct {
	baseURL, user, pass string
}

var rootCmd = &cobra.Command{
	Use:   "transfersh [file|directory]",
	Short: "Transfersh files or directories.",
	Long: `transfersh is a CLI tool for uploading files or directories.
Given a file or directory path, it will upload the content and 
provide a URL for download. If a directory path is provided,
it will be compressed as a .zip file and then uploaded.`,
	Args: cobra.ExactArgs(1),
	Run:  executeTransfer,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeTransfer(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	file := args[0]
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("%s: No such file or directory\n", file)
		return
	}

	fileName := filepath.Base(file)
	reader, size, err := prepareContent(file)
	if err != nil {
		fmt.Println("Error preparing content:", err)
		return
	}

	if info, err := os.Stat(file); err == nil && info.IsDir() {
		fileName += ".zip"
	}

	uploadAndPrint(reader, fileName, config, size)
}

// loadConfig loads the config file
func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".config", "transfersh", ".config")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return createConfig(configPath)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("invalid config format")
	}

	return &Config{
		baseURL: lines[0],
		user:    lines[1],
		pass:    lines[2],
	}, nil
}

// createConfig creates a new config file
func createConfig(path string) (*Config, error) {
	dir := filepath.Dir(path)                      // get directory
	if err := os.MkdirAll(dir, 0755); err != nil { // create directory
		return nil, err
	}

	var config Config // create config
	fmt.Print("Enter transfer base URL: ")
	if _, err := fmt.Scanln(&config.baseURL); err != nil { // get base URL
		return nil, err
	}
	fmt.Print("Enter transfer user: ")
	if _, err := fmt.Scanln(&config.user); err != nil { // get user
		return nil, err
	}
	fmt.Print("Enter transfer pass: ")
	if _, err := fmt.Scanln(&config.pass); err != nil { // get pass
		return nil, err
	}

	content := fmt.Sprintf("%s\n%s\n%s", config.baseURL, config.user, config.pass) // create content
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {              // write content
		return nil, err
	}

	return &config, nil
}

// prepareContent prepares the content for upload
func prepareContent(file string) (io.Reader, int64, error) {
	info, err := os.Stat(file) // get file info
	if info.IsDir() {
		reader, size, err := zipDirectory(file)
		return reader, size, err
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, 0, err
	}
	fileInfo, _ := f.Stat()
	return f, fileInfo.Size(), nil
}

// zipDirectory zips the directory and returns a reader
func zipDirectory(directory string) (io.Reader, int64, error) {
	var buf bytes.Buffer             // create buffer
	zipWriter := zip.NewWriter(&buf) // create zip writer

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil { // skip errors
			return err
		}
		if info.IsDir() { // skip directories
			return nil
		}
		relPath, _ := filepath.Rel(directory, path) // get relative path
		zf, _ := zipWriter.Create(relPath)          // create zip file
		content, err := os.ReadFile(path)           // read file content
		if err != nil {
			return err
		}
		_, err = zf.Write(content) // write file content to zip file
		return err
	})

	if err != nil {
		return nil, 0, err
	}

	if err = zipWriter.Close(); err != nil { // close zip writer
		return nil, 0, err
	}

	return &buf, int64(buf.Len()), nil
}

func uploadContent(reader io.Reader, fileName string, config *Config, size int64) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", config.baseURL+"/"+fileName, reader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	bar := pb.Full.Start64(size)
	barReader := bar.NewProxyReader(reader)
	req.Body = io.NopCloser(barReader)
	req.ContentLength = size

	req.SetBasicAuth(config.user, config.pass)
	resp, err := client.Do(req)
	bar.Finish()

	if err != nil {
		return nil, fmt.Errorf("error uploading: %w", err)
	}
	return resp, nil
}

func printResponse(resp *http.Response, size int64, config *Config, fileName string) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload: %s", body)
	}

	// Copy to clipboard if possible, don't fail if it's not possible
	if err := clipboard.WriteAll(string(body)); err != nil {
		fmt.Println("Error copying to clipboard:", err)
	}

	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)

	sizeStr := fmtSize(size)
	_, err = fmt.Fprintf(w, "\n%s (%s):\t", blue(fileName), yellow(sizeStr))
	if err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}
	_, err = fmt.Fprintf(w, "\n%s\t%s\t", green("File URL:"), body)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}

	directDownloadLink := strings.Replace(string(body), config.baseURL+"/", config.baseURL+"/get/", 1)
	_, err = fmt.Fprintf(w, "\n%s\t%s\t", green("Direct Download Link:"), directDownloadLink)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}

	for name, values := range resp.Header {
		if strings.ToLower(name) == "x-url-delete" {
			_, err = fmt.Fprintf(w, "\n%s\tcurl -X DELETE \"%s\"\t", red("To delete use:"), values[0])
			if err != nil {
				return fmt.Errorf("error writing to stdout: %w", err)
			}
		}
	}

	_, err = fmt.Fprintln(w)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}
	err = w.Flush()
	if err != nil {
		return fmt.Errorf("error flushing stdout: %w", err)
	}

	return nil
}

func uploadAndPrint(reader io.Reader, fileName string, config *Config, size int64) {
	resp, err := uploadContent(reader, fileName, config, size)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = printResponse(resp, size, config, fileName)
	if err != nil {
		fmt.Println(err)
	}
}

func fmtSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%d KB", size/1024)
	}
	if 1024*1024 <= size && size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/float64(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/float64(1024*1024*1024))
}
