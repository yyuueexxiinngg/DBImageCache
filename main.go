package main

import (
	"DBImageCache/config"
	"DBImageCache/file"
	"DBImageCache/jav"
	"DBImageCache/logger"
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-ieproxy"
	"github.com/unrolled/secure"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed  static
var static embed.FS

var javStore = jav.JavStore{Repeat: 3, Limit: make(chan struct{}, 10)}
var javBest = jav.JavBest{Repeat: 3, Limit: make(chan struct{}, 10)}
var javScreens = jav.JavScreens{Repeat: 3, Limit: make(chan struct{}, 10)}
var javPop = jav.JavPop{Repeat: 3, Limit: make(chan struct{}, 10)}

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

//SearchJavbest
//SearchJavstore 部分被删除图片
//javScreen  电影少
//javpop 图片较小
func main() {
	javBest.Search("REBD-447")
	var history sync.Map
	r := gin.Default()

	r.Use(TlsHandler())
	//r.Use(timeoutMiddleware(time.Second * 60))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/img/:javID", func(c *gin.Context) {
		javID := c.Param("javID")
		isFc2 := strings.HasPrefix(javID, "FC2-")
		if isFc2 {
			javID = "FC2-PPV-" + javID[4:]
		}
		LoadImg := func() {
			if file.IsExist(config.ImgPath() + javID + ".jpg") {
				c.File(config.ImgPath() + javID + ".jpg")
				return
			}
			c.FileFromFS("static/notFound.png", http.FS(static))
		}

		//start
		if running, ok := history.Load(javID); ok {
			<-(*running.(*chan struct{}))
			LoadImg()
			return
		}

		//防止相同的javID进来，重新查询
		done := make(chan struct{}, 1)
		defer close(done)
		history.Store(javID, &done)

		if file.IsExist(config.ImgPath() + javID + ".jpg") {
			c.File(config.ImgPath() + javID + ".jpg")
			return
		}
		//todo: 记录每个番号的访问时间，磁盘满10G时，删除最久没访问过的100M文件
		if err := javStore.Download(javID); errors.Is(err, jav.ErrNotFound) {
			logger.Info("JavStore Not Found: " + javID)
		} else if err != nil {
			logger.Error("Jav [" + javID + "]: " + err.Error())
		} else {
			logger.Info("JavStore Found: " + javID)
			c.File(config.ImgPath() + javID + ".jpg")
			return
		}

		if err := javBest.Download(javID); errors.Is(err, jav.ErrNotFound) {
			logger.Info("JavBest Not Found: " + javID)
		} else if err != nil {
			logger.Error("javBest [" + javID + "]: " + err.Error())
		} else {
			logger.Info("JavBest Found: " + javID)
			c.File(config.ImgPath() + javID + ".jpg")
			return
		}

		if err := javScreens.Download(javID); errors.Is(err, jav.ErrNotFound) {
			logger.Info("JavScreens Not Found: " + javID)
		} else if err != nil {
			logger.Error("JavScreens [" + javID + "]: " + err.Error())
		} else {
			logger.Info(javID + "JavScreens Found: " + javID)
			c.File(config.ImgPath() + javID + ".jpg")
			return
		}

		if err := javPop.Download(javID); errors.Is(err, jav.ErrNotFound) {
			logger.Info("javPop Not Found:" + javID)
		} else if err != nil {
			logger.Error("javPop [" + javID + "]: " + err.Error())
		} else {
			logger.Info("javPop Found: " + javID)
			c.File(config.ImgPath() + javID + ".jpg")
			return
		}
		logger.Info(javID + "：Not Found")
		c.FileFromFS("static/notFound.png", http.FS(static))
	})

	r.RunTLS("127.0.0.1:443", "./localhost.crt", "./localhost.key")
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
