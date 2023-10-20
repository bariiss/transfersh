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
