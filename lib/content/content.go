package content

import (
	"fmt"
	"github.com/bariiss/transfersh/lib"
	c "github.com/bariiss/transfersh/lib/config"
	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Path: lib/content/content.go

var MaxDays string
var MaxDownloads string

// PrepareContent prepares the content for uploading
func PrepareContent(filePath string) (fileName string, reader io.Reader, size int64, err error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return
	}

	if fileInfo.IsDir() {
		fileName = filepath.Base(filePath) + ".zip"
		zipPath := filepath.Join(os.TempDir(), fileName)
		err = ZipDirectory(filePath, zipPath)
		if err != nil {
			return
		}
		reader, err = os.Open(zipPath)
		if err != nil {
			return
		}
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Println("Error removing zip file:", err)
			}
		}(zipPath) // ensure zip file is removed after uploading
		size = fileInfo.Size()
		return
	}

	fileName = filepath.Base(filePath)
	reader, err = os.Open(filePath)
	size = fileInfo.Size()
	return
}

func UploadContent(fileName string, reader io.Reader, size int64, config *c.Config, maxDays string, maxDownloads string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("PUT", config.BaseURL+"/"+fileName, reader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// add some headers -H "Max-Days: 1" -H "Max-Downloads: 1"
	req.Header.Add("Max-Days", maxDays)
	req.Header.Add("Max-Downloads", maxDownloads)

	bar := pb.Full.Start64(size)
	barReader := bar.NewProxyReader(reader)
	req.Body = io.NopCloser(barReader)
	req.ContentLength = size

	req.SetBasicAuth(config.User, config.Pass)
	resp, err := client.Do(req)
	bar.Finish()

	if err != nil {
		return nil, fmt.Errorf("error uploading: %w", err)
	}
	return resp, nil
}

// ExecuteTransfer executes the transfer command
func ExecuteTransfer(cmd *cobra.Command, args []string) {
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

	fileName, reader, size, err := PrepareContent(file)
	if err != nil {
		fmt.Println("Error preparing content:", err)
		return
	}

	if info, err := os.Stat(file); err == nil && info.IsDir() { // check if file is directory
		fileName += ".zip" // add .zip extension
	}

	resp, err := UploadContent(fileName, reader, size, loadConfig, MaxDays, MaxDownloads)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = lib.PrintResponse(resp, size, loadConfig, fileName)
	if err != nil {
		fmt.Println(err)
	}
}
