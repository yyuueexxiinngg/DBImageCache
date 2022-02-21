package jav

import (
	"DBImageCache/config"
	"DBImageCache/file"
	"DBImageCache/utils"
	"bytes"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
)

type BlogJav struct {
	Repeat int
	Limit  chan struct{}
}

func (j BlogJav) Download(javID string) error {
	j.Limit <- struct{}{}
	defer func() { <-j.Limit }()

	var url string
	var err error
	for t := 0; t < j.Repeat; t++ {
		url, err = j.Search(javID)
		//遇到网络错误进行重试
		if err != nil {
			if t == j.Repeat-1 {
				return err
			}
			continue
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

func (j BlogJav) Search(javID string) (string, error) {
	if strings.HasPrefix(javID, "FC2") {
		javID = strings.ReplaceAll(javID, "-", " ")
	}

	content, err := utils.GetWithTime("http://blogjav.net/?s="+javID, connectTime)
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

	var selc *goquery.Selection
	var reg = regexp.MustCompile(`\b` + titleContent + `\b`)
	doc.Find(".content-area .inside-article h2 a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if title := s.Text(); reg.MatchString(title) {
			title = strings.ToLower(title)
			if selc == nil {
				selc = s
			}
			if strings.Contains(title, "uncensored") {
				selc = s
				return false
			} else if strings.Contains(title, "4k") || strings.Contains(title, "fhd") {
				selc = s
			}
		}
		return true
	})
	if selc == nil {
		return "", ErrNotFound
	}

	//fmt.Println(selc.Attr("href"))

	detailPage, ok := selc.Attr("href")
	if !ok {
		return "", errors.New("not found href")
	}

	content, err = utils.GetWithTime(detailPage, connectTime)
	if err != nil {
		return "", err
	}

	doc2, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	url, ok := doc2.Find(".entry-content a").First().Find("img").Attr("data-lazy-src")
	if !ok || url == "" {
		return "", ErrNotFound
	}
	url = strings.ReplaceAll(url, "pixhost.org", "pixhost.to")
	url = strings.ReplaceAll(url, ".th", "")
	url = strings.ReplaceAll(url, "thumbs", "images")
	url = strings.ReplaceAll(url, "//t", "//img")
	re3, _ := regexp.Compile("[\\?*\\\"*]")
	url = re3.ReplaceAllString(url, "")

	return url, nil
}
