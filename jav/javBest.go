package jav

import (
	"DBImageCache/config"
	"DBImageCache/file"
	"DBImageCache/utils"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
)

type JavBest struct {
	Repeat int
	Limit  chan struct{}
}

func (j JavBest) Download(javID string) error {
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
		} else {
			break
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

func (j JavBest) Search(javID string) (string, error) {
	content, err := utils.GetWithTime("http://javbest.net/?s="+javID, connectTime)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	titleContent := javID
	if strings.HasPrefix(javID, "FC2") {
		_, titleContent = utils.SplitJavID(javID)
	}

	var reg = regexp.MustCompile(`\b` + titleContent + `\b`)

	// Find the review items
	var selc *goquery.Selection
	doc.Find(".app_main_container article h2 a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if title, ok := s.Attr("title"); ok && reg.MatchString(title) {
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

	searchPage, ok := selc.Attr("href")
	if !ok {
		panic("not found href")
		//return "", errors.New("not found href")
	}

	content, err = utils.GetWithTime(searchPage, connectTime)
	if err != nil {
		return "", err
	}
	doc2, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	detailPage, ok := doc2.Find(".post-single-content p a").First().Attr("href")
	if !ok || detailPage == "" {
		panic("not found href")
		//return "", errors.New("not found href")
	}

	content, err = utils.GetWithTime(detailPage, connectTime)
	if err != nil {
		return "", err
	}
	doc3, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	url, ok := doc3.Find(".pic").First().Attr("src")
	if !ok || url == "" {
		return "", ErrNotFound
	}

	return url, nil
}
