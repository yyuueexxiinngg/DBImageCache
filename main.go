package main

import (
	"DBImageCache/file"
	"DBImageCache/jav"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-ieproxy"
	"github.com/unrolled/secure"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
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

	//检测证书并安装
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

const (
	WaitDownload = iota
	Fine
)

var notFount sync.Map

//SearchJavbest
//SearchJavstore 部分被删除图片
//javscreen  电影少
//javpop 图片较小
func main() {
	r := gin.Default()
	//r.Use(timeoutMiddleware(time.Second * 60))
	r.Use(TlsHandler())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/img/:javID", func(c *gin.Context) {
		javID := c.Param("javID")
		//ctx := c.Request.Context()

		if status, ok := jav.JavImageLocalFiles.Load(javID); ok {
			<-status.(<-chan struct{})
			c.File("./static/" + javID + ".jpg")
			return
		}
		if file.IsExist("./static/" + javID + ".jpg") {
			c.File("./static/" + javID + ".jpg")
			return
		}
		if _, ok := notFount.Load(javID); ok {
			c.File("./static/notFound.png")
			return
		}

		//todo:查找提示网络错误，还是真的没找到
		//jav
		if jav.JavStore(javID) {
			c.File("./static/" + javID + ".jpg")
			return
		}
		if jav.JavBest(javID) {
			c.File("./static/" + javID + ".jpg")
			return
		}
		if jav.JavScreens(javID) {
			c.File("./static/" + javID + ".jpg")
			return
		}
		if jav.JavPop(javID) {
			c.File("./static/" + javID + ".jpg")
			return
		}
		notFount.Store(javID, true)
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

func timeoutMiddleware(timeout time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)

		defer func() {
			// check if context timeout was reached
			if ctx.Err() == context.DeadlineExceeded {

				// write response and abort the request
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			}

			//cancel to clear resources after finished
			cancel()
		}()

		// replace request with context wrapped request
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
