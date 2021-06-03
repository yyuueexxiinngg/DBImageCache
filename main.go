package main

import (
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"net/http"
)


func main(){
	r := gin.Default()
	r.Use(TlsHandler())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.StaticFS("/static", http.Dir("./static"))
	r.RunTLS(":8080", "./server.crt","./server.key")
}

func TlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "localhost:8080",
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}