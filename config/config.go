package config

import (
	"io/fs"
	"os"
	"path"
	"runtime"
)

var staticPath string

func init() {
	//图片保存路径win默认：%ALLUSERSPROFILE%/DBImageCache
	if runtime.GOOS == "windows" {
		staticPath = path.Join(os.Getenv("ALLUSERSPROFILE"), "DBImageCache") + "/"
		os.MkdirAll(staticPath, fs.ModePerm)
	} else {
		staticPath = "./static/"
	}

	os.MkdirAll(staticPath+"temp/", os.ModePerm)
}

func ImgPath() string {
	return staticPath
}
