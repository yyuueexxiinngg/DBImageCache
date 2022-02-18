package jav

import (
	"DBImageCache/config"
	"DBImageCache/logger"
	"errors"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var downloadTime = 60 * time.Second
var connectTime = 20 * time.Second

//todo: 屏蔽VR相关，减少查找
var VRLists = []string{
	"CVPS",
	"DSVR",
	"EIN",
	"HNVR",
	"IPVR",
	"JUVR",
	"KAVR",
	"OVVR",
	"VRKM",
	"WAVR",
	"MDVR",
}

type JavImger interface {
	Search() (url string, err error)
}

var (
	ErrNotFound       = errors.New("jav not found")
	ErrDownloadFailed = errors.New("jav download failed")
)

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

func DownloadImage(url, filePath, javID string) *errgroup.Group {
	//done := make(chan error, 1)
	var g errgroup.Group

	fileName := javID + ".jpg"
	g.Go(func() error {

		client := http.Client{Timeout: downloadTime}

		logger.Info("Start to download: " + url)
		response, err := client.Get(url)
		if err != nil {
			return err
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			if response.StatusCode == http.StatusNotFound {
				return ErrNotFound
			} else {
				return ErrDownloadFailed
			}
		}

		contentLength := response.ContentLength
		if contentLength < 30000 {
			return ErrNotFound
		}
		logger.Info("ContentLength: " + strconv.FormatInt(contentLength, 10))
		if contentLength != SaveImage(config.ImgPath()+"temp/"+fileName, response.Body) {
			//放弃临时文件夹的文件
			logger.Info("放弃临时文件夹的文件")
			os.Remove(config.ImgPath() + "temp/" + fileName)
			return nil
		}
		//将临时文件夹的文件复制的static里
		err = os.Rename(config.ImgPath()+"temp/"+fileName, filePath+fileName)
		if err != nil {
			logger.Error(fileName + " file move error: " + err.Error())
			return err
		}
		return nil
	})
	return &g
}
