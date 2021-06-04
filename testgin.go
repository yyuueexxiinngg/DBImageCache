package main

import (
	"DBImageCache/file"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/:id", func(c *gin.Context) {
		name := c.Param("id")

		if file.IsExist("./static/" + name) {
			c.File("./static/" + name )
		}

		// 下载图片
		file.DownloadImage("https://static.studygolang.com/static/img/logo.png","./static/1")
		//https://cdn.pixabay.com/photo/2021/04/22/04/09/rapeseed-6197976__340.jpg

		c.File("./static/" + name )

		//c.String(http.StatusNotFound, "没有这张图片")

	})
	r.Run() // listen and serve on 0.0.0.0:8080
}