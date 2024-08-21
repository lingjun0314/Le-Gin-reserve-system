package main

import (
	"LeGinReserve/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	routers.InitUserRouter(r)
	
	r.Run()
}
