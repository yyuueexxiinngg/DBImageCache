package main

import (
	"DBImageCache/file"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-ieproxy"
	"github.com/unrolled/secure"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func init() {
	//获取系统代理
	http.DefaultTransport.(*http.Transport).Proxy = ieproxy.GetProxyFunc()

	//设置运行文件所在目录为当前目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}
	if err = os.Chdir(dir); err != nil {
		log.Fatalln(err)
	}

	//检测证书
	if !file.IsExist("localhost.crt") {
		cmd := exec.Command("./mkcert_win/mkcert.exe", "-install")
		if result, err := cmd.Output(); err != nil {
			fmt.Println(err, result)
			os.Exit(1)
		}
		cmd = exec.Command("./mkcert_win/mkcert.exe", "-cert-file", "./localhost.crt", "-key-file", "./localhost.key", "114.taobao.com", "127.0.0.1", "::1")
		if result, err := cmd.Output(); err != nil {
			fmt.Println(err, result)
			os.Exit(1)
		}
	}

}

var javImageLocalFiles sync.Map

const (
	WaitDownload = iota
	Fine
)

//SearchJavbest
//SearchJavstore 部分被删除图片
//javscreen  电影少
//javpop 图片较小
func main() {
	//url, ok := SearchJavstore("SSIS-080")
	//fmt.Println(url, ok)
	//fmt.Println(SearchJavbest("300NTK-181"))

	r := gin.Default()
	r.Use(TlsHandler())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//r.GET("/download/:javID", func(c *gin.Context) {
	//	defer c.Status(200)
	//	javID := c.Param("javID")
	//	_, ok := javImageLocalFiles.Load(javID)
	//	if ok {
	//		return
	//	}
	//
	//	if file.IsExist("./static/" + javID + ".jpg") {
	//		return
	//	}
	//	//todo:网站不存在此图片，缓存提示，不要重复去查找
	//	//done := file.DownloadImage("http://javScreens.com/images/"+javID+".jpg", "./static/"+javID+".jpg")
	//	//javImageLocalFiles.Store(javID, done)
	//	//<-done
	//	//javImageLocalFiles.Delete(javID)
	//	javStore(javID)
	//	javScreens(javID)
	//})

	r.GET("/img/:javID", func(c *gin.Context) {
		javID := c.Param("javID")

		status, ok := javImageLocalFiles.Load(javID)
		if ok {
			<-status.(<-chan struct{})
			c.File("./static/" + javID + ".jpg")
			return
		}

		if file.IsExist("./static/" + javID + ".jpg") {
			c.File("./static/" + javID + ".jpg")
			return
		}
		//请求文件下载

		//todo: 从多个网站寻找:  3xplanet.com  ||     akiba-online.com   javfree.me

		//javStore
		if javStore(javID) {
			c.File("./static/" + javID + ".jpg")
			return
		}

		if javBest(javID) {
			c.File("./static/" + javID + ".jpg")
			return
		}

		//javScreens
		if javScreens(javID) {
			c.File("./static/" + javID + ".jpg")
			return
		}

		if javPop(javID) {
			c.File("./static/" + javID + ".jpg")
			return
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

func javBest(javID string) bool {
	url, ok := SearchJavbest(javID)
	if !ok {
		return false
	}
	done := file.DownloadImage(url, "./static/"+javID+".jpg")
	javImageLocalFiles.Store(javID, done)
	<-done
	javImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		//c.File("./static/" + javID + ".jpg")
		return true
	}
	return false
}

func javStore(javID string) bool {
	url, ok := SearchJavstore(javID)
	if !ok {
		return false
	}
	done := file.DownloadImage(url, "./static/"+javID+".jpg")
	javImageLocalFiles.Store(javID, done)
	<-done
	javImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
		//c.File("./static/" + javID + ".jpg")
		return true
	}
	return false
}

func javPop(javID string) bool {
	url, ok := SearchJavpop(javID)
	if !ok {
		return false
	}

	done := file.DownloadImage(url, "./static/"+javID+".jpg")
	javImageLocalFiles.Store(javID, done)
	<-done
	javImageLocalFiles.Delete(javID)
	if file.IsExist("./static/" + javID + ".jpg") {
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

func SearchJavpop(javID string) (string, bool) {
	//javID := "300NTK-181"

	res, err := http.Get("http://javpop.com/index.php?s=" + javID)
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

	url, ok := doc2.Find(".screenshot img").First().Attr("src")
	if !ok || url == "" {
		return "", false
	}

	return url, true
}

func SearchJavbest(javID string) (string, bool) {
	//javID := "300NTK-181"

	res, err := http.Get("http://javbest.net/?s=" + javID)
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

	res2, err := http.Get(searchPage)
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

	res3, err := http.Get(detailPage)
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
