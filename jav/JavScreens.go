package jav

import (
	"DBImageCache/config"
	"DBImageCache/file"
	"strings"
)

type JavScreens struct {
	Repeat int
	Limit  chan struct{}
}

func (j JavScreens) Download(javID string) error {
	j.Limit <- struct{}{}
	defer func() { <-j.Limit }()
	if strings.HasPrefix(javID, "FC2-PPV") {
		return ErrNotFound
	}
	//todo: download repeat try
	g := DownloadImage("http://javScreens.com/images/"+javID+".jpg", config.ImgPath(), javID)
	err := g.Wait()
	if err != nil {
		return err
	}
	if file.IsExist(config.ImgPath() + javID + ".jpg") {
		return nil
	}
	return ErrDownloadFailed
}
