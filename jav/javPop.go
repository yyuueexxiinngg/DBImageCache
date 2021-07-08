package jav

import (
	"DBImageCache/config"
	"DBImageCache/file"
	"DBImageCache/utils"
	"bytes"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

type JavPop struct {
	Repeat int
	Limit  chan struct{}
}

func (j JavPop) Download(javID string) error {
	j.Limit <- struct{}{}
	defer func() { <-j.Limit }()

	var url string
	var err error
	for t := 0; t < j.Repeat; t++ {
		url, err = j.Search(javID)
		//遇到网络错误进行重试
		if err != nil {
			if t != j.Repeat-1 {
				continue
			}
			return err
		}
		//没找到直接返回
		if url == "" {
			return ErrNotFound
		}
	}

	g := DownloadImage(url, config.ImgPath(), javID)
	err = g.Wait()
	if err != nil {
		return err
	}
	if file.IsExist(config.ImgPath() + javID + ".jpg") {
		return nil
	}
	return ErrDownloadFailed
}

func (j JavPop) Search(javID string) (string, error) {
	if strings.HasPrefix(javID, "FC2-PPV-") {
		javID = strings.ReplaceAll(javID, "FC2-PPV-", "FC2_PPV-")
	}

	content, err := utils.GetWithTime("http://javpop.com/index.php?s="+javID, connectTime)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(content))
	if err != nil {
		return "", err
	}

	// Find the review items
	var selc *goquery.Selection
	doc.Find(".entry a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if title, ok := s.Attr("title"); ok && strings.Contains(title, "["+javID+"]") {
			if strings.Contains(title, "Uncensored") || strings.Contains(title, "FHD") {
				selc = s
				return false
			}
			selc = s
		}
		return true
	})
	if selc == nil {
		return "", ErrNotFound
	}

	detailPage, ok := selc.Attr("href")
	if !ok {
		return "", errors.New("not href")
	}

	content, err = utils.GetWithTime(detailPage, connectTime)

	doc2, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	url, ok := doc2.Find(".screenshot img").First().Attr("src")
	if !ok || url == "" {
		return "", errors.New("not screenshot img")
	}

	return url, nil
}
