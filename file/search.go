package file

import (
	"os"
)

func IsExist(filePath string) bool {
	f, err := os.Stat(filePath) //os.Stat获取文件信息
	return err == nil || os.IsExist(err) && !f.IsDir()
}
