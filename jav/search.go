package jav

import (
	"DBImageCache/file"
	"DBImageCache/logger"
	"DBImageCache/utils"
	"bytes"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var JavImageLocalFiles sync.Map

var connectTime = 10 * time.Second

func init() {
	err := os.MkdirAll("./temp", os.ModePerm)
	if err != nil {
		logger.Error(err.Error())
		return
	}
}

func JavScreens(javID string) bool {
	done := file.DownloadImage("http://javScreens.com/images/"+javID+".jpg", "./static/", javID+".jpg")
	JavImageLocalFiles.Store(javID, done)
	<-done
	JavImageLocalFiles.Delete(javID)

	if file.IsExist("./static/" + javID + ".jpg") {
		return true
	}
	return false
}

func JavBestSearch(searchID, javID string) bool {
	url, err := SearchJavbest(searchID)
	if err != nil {
		return false
	}
	done := file.DownloadImage(url, "./static/", javID+".jpg")
	JavImageLocalFiles.Store(javID, done)
	<-done
	JavImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		//c.File("./static/" + javID + ".jpg")
		return true
	}
	return false
}

func JavBest(javID string) bool {
	url, err := SearchJavbest(javID)
	if err != nil {
		return false
	}
	done := file.DownloadImage(url, "./static/", javID+".jpg")
	JavImageLocalFiles.Store(javID, done)
	<-done
	JavImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		//c.File("./static/" + javID + ".jpg")
		return true
	}
	return false
}

func JavStoreSearch(searchID, javID string) bool {
	url, err := SearchJavstore(searchID)
	if err != nil {
		return false
	}
	done := file.DownloadImage(url, "./static/", javID+".jpg")
	JavImageLocalFiles.Store(javID, done)
	<-done
	JavImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		//c.File("./static/" + javID + ".jpg")
		return true
	}
	return false
}

func JavStore(javID string) bool {
	url, err := SearchJavstore(javID)
	if err != nil {
		return false
	}
	done := file.DownloadImage(url, "./static/", javID+".jpg")
	JavImageLocalFiles.Store(javID, done)
	<-done
	JavImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		//c.File("./static/" + javID + ".jpg")
		return true
	}
	return false
}

func JavPopSearch(searchID, javID string) bool {
	url, err := SearchJavpop(searchID)
	if err != nil {
		return false
	}

	done := file.DownloadImage(url, "./static/", javID+".jpg")
	JavImageLocalFiles.Store(javID, done)
	<-done
	JavImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		return true
	}
	return false
}

func JavPop(javID string) bool {
	url, err := SearchJavpop(javID)
	if err != nil {
		return false
	}

	done := file.DownloadImage(url, "./static/", javID+".jpg")
	JavImageLocalFiles.Store(javID, done)
	<-done
	JavImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		return true
	}
	return false
}

func SearchJavstore(javID string) (string, error) {
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
	doc.Find(".news_1n li h3 span a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if title, ok := s.Attr("title"); ok && strings.Contains(title, titleContent) {
			if strings.Contains(title, "Uncensored") || strings.Contains(title, "FHD") {
				selc = s
				return false
			}
			selc = s
		}
		return true
	})
	if selc == nil {
		return "", errors.New("not found content")
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
		return "", errors.New("not found img")
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

func SearchJavpop(javID string) (string, error) {
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
		if title, ok := s.Attr("title"); ok && strings.Contains(title, javID) {
			if strings.Contains(title, "Uncensored") || strings.Contains(title, "FHD") {
				selc = s
				return false
			}
			selc = s
		}
		return true
	})
	if selc == nil {
		return "", errors.New("not found content")
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

func SearchJavbest(javID string) (string, error) {
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

	// Find the review items
	var selc *goquery.Selection
	doc.Find(".content-area article h1 a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if title, ok := s.Attr("title"); ok && strings.Contains(title, titleContent) {
			if strings.Contains(title, "Uncensored") || strings.Contains(title, "FHD") {
				selc = s
				return false
			}
			selc = s
		}
		return true
	})
	if selc == nil {
		return "", errors.New("can't search")
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

	detailPage, ok := doc2.Find(".entry-content p a").First().Attr("href")
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
		return "", errors.New("not found img")
	}

	return url, nil
}
