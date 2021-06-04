package file

import (
	"io"
	"log"
	"net/http"
	"os"
)

func IsExist(filePath string) bool {
	f, err := os.Stat(filePath) //os.Stat获取文件信息
	return err == nil || os.IsExist(err) && !f.IsDir()
}

func SaveImage(filePath string, content io.Reader) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, content)
	if err != nil {
		log.Fatal(err)
	}
}

func DownloadImage( url, filePath string) {
	// don't worry about errors
	response, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()
	SaveImage(filePath, response.Body)
}
