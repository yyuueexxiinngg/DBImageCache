package main

import (
	"DBImageCache/file"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-ieproxy"
	"github.com/unrolled/secure"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

func init() {
	http.DefaultTransport.(*http.Transport).Proxy = ieproxy.GetProxyFunc()
}

var javImageLocalFiles sync.Map

const (
	WaitDownload = iota
	Fine
)

func main() {
	//url, ok := SearchJavstore("SSIS-080")
	//fmt.Println(url, ok)

	r := gin.Default()
	r.Use(TlsHandler())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/download/:javID", func(c *gin.Context) {
		defer c.Status(200)
		javID := c.Param("javID")
		_, ok := javImageLocalFiles.Load(javID)
		if ok {
			return
		}

		if file.IsExist("./static/" + javID + ".jpg") {
			return
		}
		//todo:网站不存在此图片，缓存提示，不要重复去查找
		//done := file.DownloadImage("http://javScreens.com/images/"+javID+".jpg", "./static/"+javID+".jpg")
		//javImageLocalFiles.Store(javID, done)
		//<-done
		//javImageLocalFiles.Delete(javID)
		javStore(javID)
		javScreens(javID)
	})

	r.GET("/img/:javID", func(c *gin.Context) {
		javID := c.Param("javID")

		status, ok := javImageLocalFiles.Load(javID)
		if ok {
			<-status.(<-chan struct{})
			c.File("./static/" + javID + ".jpg")
		}

		if file.IsExist("./static/" + javID + ".jpg") {
			c.File("./static/" + javID + ".jpg")
		}
		//请求文件下载

		//todo: 从多个网站寻找

		//javStore
		if javStore(javID) {
			c.File("./static/" + javID + ".jpg")
		}

		//javScreens
		if javScreens(javID) {
			c.File("./static/" + javID + ".jpg")
		}

		c.File("./static/notFound.png")
	})

	//r.StaticFS("/static", http.Dir("./static"))
	r.RunTLS(":443", "./localhost.crt", "./localhost.key")
	//r.Run(":8080")
}

func TlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "114.taobao.com:443",
		})
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			return
		}

		c.Next()
	}
}

func javScreens(javID string) bool {
	done := file.DownloadImage("http://javScreens.com/images/"+javID+".jpg", "./static/"+javID+".jpg")
	javImageLocalFiles.Store(javID, done)
	<-done
	javImageLocalFiles.Delete(javID)

	if file.IsExist("./static/" + javID + ".jpg") {
		return true
	}
	return false
}

func javStore(javID string) bool {
	url, ok := SearchJavstore(javID)
	if !ok {
		return false
	}
	done2 := file.DownloadImage(url, "./static/"+javID+".jpg")
	javImageLocalFiles.Store(javID, done2)
	<-done2
	javImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		//c.File("./static/" + javID + ".jpg")
		return true
	}
	return false
}

func SearchJavstore(javID string) (string, bool) {
	//javID := "OREC-769"
	//javID := "SSIS-086"
	res, err := http.Get("http://javStore.net/search/" + javID + ".html")
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
	selc := doc.Find(".news_1n li h3 span a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// For each item found, get the title
		if title, ok := s.Attr("title"); ok && strings.Contains(title, javID) {
			//fmt.Println(s.Attr("href"))
			if strings.Contains(title, "Uncensored") || strings.Contains(title, "FHD") {
				return false
			}

		}
		return true
	})
	//fmt.Println(selc.Attr("href"))

	detailPage, ok := selc.Attr("href")
	if !ok {
		return "", false
	}

	res2, err := http.Get(detailPage)
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
