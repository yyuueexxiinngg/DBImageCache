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

type JavStore struct {
	Repeat int
	Limit  chan struct{}
}

func (j JavStore) Download(javID string) error {
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

func (j JavStore) Search(javID string) (string, error) {
	content, err := utils.GetWithTime("http://javStore.net/search/"+javID+".html", connectTime)
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
	doc.Find(".news_1n li h3 span a").EachWithBreak(func(i int, s *goquery.Selection) bool {
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
	url := ""
	doc2.Find(".news a img[alt*=th]").Each(func(i int, s *goquery.Selection) {
		//fmt.Println(s.Attr("src"))
		url, ok = s.Attr("src")
	})
	if !ok || url == "" {
		return "", ErrNotFound
	}
	//url := "https://img.javstore.net/images/2021/06/03/SSIS-086_s.th.jpg"
	url = strings.ReplaceAll(url, "pixhost.org", "pixhost.to")
	url = strings.ReplaceAll(url, ".th", "")
	url = strings.ReplaceAll(url, "thumbs", "images")
	url = strings.ReplaceAll(url, "//t", "//img")
	re3, _ := regexp.Compile("[\\?*\\\"*]")
	url = re3.ReplaceAllString(url, "")
	url = strings.ReplaceAll(url, "https", "http")

	return url, nil
}
