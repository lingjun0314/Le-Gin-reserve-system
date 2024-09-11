package main

import (
	"LeGinReserve/routers"
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	//	Set releaseMode
	gin.SetMode(gin.ReleaseMode)

	//	Disable console color
	gin.DisableConsoleColor()

	//	Create log file
	f, err := os.Create("gin.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	//	Set log file write in file and console
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	r := gin.Default()

	//	Set trust proxy ip
	r.SetTrustedProxies([]string{"127.0.0.1"})
	routers.InitUserRouter(r)

	r.Run(":8080")
}
