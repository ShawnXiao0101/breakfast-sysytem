package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/order", createOrder)
	r.GET("/orders", getOrders)
	r.GET("/ws", serveWS)
	r.POST("/order/:id/status", updateStatus)

	r.Run(":8080")
}
