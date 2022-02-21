package main

import (
	"DBImageCache/config"
	"DBImageCache/file"
	"DBImageCache/jav"
	"DBImageCache/logger"
	"context"
	"embed"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-ieproxy"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed  static
var static embed.FS

var blogJav = jav.BlogJav{Repeat: 3, Limit: make(chan struct{}, 10)}
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
}

//SearchJavbest
//SearchJavstore 部分被删除图片
//javScreen  电影少
//javpop 图片较小
func main() {
	//javBest.Search("REBD-447")
	var history sync.Map
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/img/:javID", func(c *gin.Context) {
		javID := strings.ToUpper(c.Param("javID"))
		isFc2 := strings.Contains(javID, "FC2")
		if isFc2 {
			re := regexp.MustCompile("[0-9]{2,}")
			javID = "FC2-PPV-" + re.FindString(javID)
		}
		logger.Info("Start getting javID: " + javID)
		LoadImg := func() {
			if file.IsExist(config.ImgPath() + javID + ".jpg") {
				c.File(config.ImgPath() + javID + ".jpg")
				return
			}
			c.JSON(404, gin.H{
				"message": "not found",
			})
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

		if err := blogJav.Download(javID); errors.Is(err, jav.ErrNotFound) {
			logger.Info("BlogJav Not Found: " + javID)
		} else if err != nil {
			logger.Error("Jav [" + javID + "]: " + err.Error())
		} else {
			logger.Info("BlogJav Found: " + javID)
			c.File(config.ImgPath() + javID + ".jpg")
			return
		}

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
		c.JSON(404, gin.H{
			"message": "not found",
		})
	})

	r.Run(config.ListenAddress)
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
