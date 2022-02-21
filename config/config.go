package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

var staticPath string
var UserAgent string
var ListenAddress string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ListenAddress = os.Getenv("LISTEN_ADDRESS")
	staticPath = os.Getenv("STATIC_PATH")
	UserAgent = os.Getenv("USER_AGENT")
	if !strings.HasSuffix(staticPath, "/") {
		staticPath += "/"
	}
	os.MkdirAll(staticPath+"/temp/", os.ModePerm)
}

func ImgPath() string {
	return staticPath
}
