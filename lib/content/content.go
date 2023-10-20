package content

import (
	"fmt"
	c "github.com/bariiss/transfersh/lib/config"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"os"
)

// Path: lib/content/content.go

// PrepareContent prepares the content for upload
func PrepareContent(file string) (io.Reader, int64, error) {
	info, err := os.Stat(file) // get file info
	if info.IsDir() {
		reader, size, err := ZipDirectory(file)
		return reader, size, err
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, 0, err
	}
	fileInfo, _ := f.Stat()
	return f, fileInfo.Size(), nil
}

func UploadContent(reader io.Reader, fileName string, config *c.Config, size int64) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", config.BaseURL+"/"+fileName, reader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

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
