package content

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// Path: lib/content/zip.go

// ZipDirectory zips the directory and returns a reader
func ZipDirectory(directory string) (io.Reader, int64, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || (info.Mode()&os.ModeSocket) != 0 {
			return err
		}

		relPath, _ := filepath.Rel(directory, path)
		zf, _ := zipWriter.Create(relPath)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = zf.Write(content)
		return err
	})

	if err != nil {
		return nil, 0, err
	}

	err = zipWriter.Close()
	return &buf, int64(buf.Len()), err
}
