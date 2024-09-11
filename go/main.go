package main

import (
	"LeGinReserve/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//	Set trust proxy ip
	r.SetTrustedProxies([]string{"127.0.0.1"})
	routers.InitUserRouter(r)

	r.Run(":8080")
}
