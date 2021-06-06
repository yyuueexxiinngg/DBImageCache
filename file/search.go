package file

import (
	"io"
	"net/http"
	"os"
	"strconv"
)

func IsExist(filePath string) bool {
	f, err := os.Stat(filePath) //os.Stat获取文件信息
	return err == nil || os.IsExist(err) && !f.IsDir()
}

func SaveImage(filePath string, content io.Reader) {
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, content)
	if err != nil {
		return
	}
}

func DownloadImage(url, filePath string) <-chan struct{} {
	done := make(chan struct{}, 1)

	go func() {
		// don't worry about errors
		defer close(done)

		response, e := http.Get(url)
		if e != nil {
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return
		}

		contentLength, _ := strconv.Atoi(response.Header.Get("Content-Length"))
		if contentLength < 30000 {
			return
		}
		SaveImage(filePath, response.Body)
	}()

	return done
}
