package lib

import (
	"fmt"
	c "github.com/bariiss/transfersh/lib/config"
	ct "github.com/bariiss/transfersh/lib/content"
	"io"
)

// Path: lib/upload.go

func Upload(reader io.Reader, fileName string, config *c.Config, size int64) {
	resp, err := ct.UploadContent(fileName, reader, size, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = PrintResponse(resp, size, config, fileName)
	if err != nil {
		fmt.Println(err)
	}
}
