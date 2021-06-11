package jav

import (
	"DBImageCache/file"
	"DBImageCache/logger"
	"github.com/PuerkitoBio/goquery"
	"net/http"
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

func JavBest(javID string) bool {
	url, ok := SearchJavbest(javID)
	if !ok {
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
	url, ok := SearchJavstore(javID)
	if !ok {
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

func JavPop(javID string) bool {
	url, ok := SearchJavpop(javID)
	if !ok {
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

func SearchJavstore(javID string) (string, bool) {
	//javID := "OREC-769"
	//javID := "SSIS-086"
	client := http.Client{
		Timeout: connectTime,
	}
	res, err := client.Get("http://javStore.net/search/" + javID + ".html")
	if err != nil {
		return "", false
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", false
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", false
	}

	// Find the review items
	selcect := doc.Find(".news_1n li h3 span a")
	var selc *goquery.Selection

	selcect.EachWithBreak(func(i int, s *goquery.Selection) bool {
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
		return "", false
	}

	//fmt.Println(selc.Attr("href"))

	detailPage, ok := selc.Attr("href")
	if !ok {
		return "", false
	}

	res2, err := client.Get(detailPage)
	if err != nil {
		return "", false
	}
	defer res2.Body.Close()

	if res2.StatusCode != http.StatusOK {
		return "", false
	}

	doc2, err := goquery.NewDocumentFromReader(res2.Body)
	if err != nil {
		return "", false
	}
	url := ""
	doc2.Find(".news a img[alt*=th]").Each(func(i int, s *goquery.Selection) {
		//fmt.Println(s.Attr("src"))
		url, ok = s.Attr("src")
	})
	if !ok || url == "" {
		return "", false
	}
	//url := "https://img.javstore.net/images/2021/06/03/SSIS-086_s.th.jpg"
	url = strings.ReplaceAll(url, "pixhost.org", "pixhost.to")
	url = strings.ReplaceAll(url, ".th", "")
	url = strings.ReplaceAll(url, "thumbs", "images")
	url = strings.ReplaceAll(url, "//t", "//img")
	re3, _ := regexp.Compile("[\\?*\\\"*]")
	url = re3.ReplaceAllString(url, "")
	url = strings.ReplaceAll(url, "https", "http")

	return url, true
}

func SearchJavpop(javID string) (string, bool) {
	//javID := "300NTK-181"
	client := http.Client{
		Timeout: connectTime,
	}
	res, err := client.Get("http://javpop.com/index.php?s=" + javID)
	if err != nil {
		return "", false
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", false
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", false
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
		return "", false
	}

	detailPage, ok := selc.Attr("href")
	if !ok {
		return "", false
	}

	res2, err := client.Get(detailPage)
	if err != nil {
		return "", false
	}
	defer res2.Body.Close()

	if res2.StatusCode != http.StatusOK {
		return "", false
	}

	doc2, err := goquery.NewDocumentFromReader(res2.Body)
	if err != nil {
		return "", false
	}

	url, ok := doc2.Find(".screenshot img").First().Attr("src")
	if !ok || url == "" {
		return "", false
	}

	return url, true
}

func SearchJavbest(javID string) (string, bool) {
	//javID := "300NTK-181"
	client := http.Client{
		Timeout: connectTime,
	}
	res, err := client.Get("http://javbest.net/?s=" + javID)
	if err != nil {
		return "", false
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", false
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", false
	}

	// Find the review items
	var selc *goquery.Selection
	doc.Find(".content-area h1 a").EachWithBreak(func(i int, s *goquery.Selection) bool {
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
		return "", false
	}

	searchPage, ok := selc.Attr("href")
	if !ok {
		return "", false
	}

	res2, err := client.Get(searchPage)
	if err != nil {
		return "", false
	}
	defer res2.Body.Close()

	if res2.StatusCode != http.StatusOK {
		return "", false
	}

	doc2, err := goquery.NewDocumentFromReader(res2.Body)
	if err != nil {
		return "", false
	}

	detailPage, ok := doc2.Find(".entry-content p a").First().Attr("href")
	if !ok || detailPage == "" {
		return "", false
	}

	res3, err := client.Get(detailPage)
	if err != nil {
		return "", false
	}
	defer res3.Body.Close()

	if res3.StatusCode != http.StatusOK {
		return "", false
	}

	doc3, err := goquery.NewDocumentFromReader(res3.Body)
	if err != nil {
		return "", false
	}

	url, ok := doc3.Find(".pic").First().Attr("src")
	if !ok || url == "" {
		return "", false
	}

	return url, true
}
