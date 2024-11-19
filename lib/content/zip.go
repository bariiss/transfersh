package content

import (
	"archive/zip"
	"os"
	"path/filepath"
)

// Path: lib/content/zip.go

// ZipDirectory zips the directory and returns a reader
func ZipDirectory(directory, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func(zipFile *os.File) {
		err := zipFile.Close()
		if err != nil {
			panic(err)
		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		err := zipWriter.Close()
		if err != nil {
			panic(err)
		}
	}(zipWriter)

	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
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
}
