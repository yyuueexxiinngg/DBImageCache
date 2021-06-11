package file

import (
	"DBImageCache/logger"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var downloadTime = 50 * time.Second

func IsExist(filePath string) bool {
	f, err := os.Stat(filePath) //os.Stat获取文件信息
	return err == nil || os.IsExist(err) && !f.IsDir()
}

func SaveImage(filePath string, content io.Reader) (written int64) {
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	// Use io.Copy to just dump the response body to the file. This supports huge files
	written, err = io.Copy(file, content)
	if err != nil {
		return
	}
	return
}

func DownloadImage(url, filePath, fileName string) <-chan struct{} {
	done := make(chan struct{}, 1)

	go func() {
		// don't worry about errors
		defer close(done)

		client := http.Client{Timeout: downloadTime}

		response, e := client.Get(url)
		if e != nil {
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return
		}

		contentLength, _ := strconv.ParseInt(response.Header.Get("Content-Length"), 10, 64)
		if contentLength < 30000 {
			return
		}
		if contentLength != SaveImage("./temp/"+fileName, response.Body) {
			//放弃临时文件夹的文件
			os.Remove("./temp/" + fileName)
			return
		}
		//将临时文件夹的文件复制的static里
		err := os.Rename("./temp/"+fileName, filePath+fileName)
		if err != nil {
			logger.Error(fileName + " file move error: " + err.Error())
			return
		}

	}()

	return done
}
