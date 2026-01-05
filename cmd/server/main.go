package main

import (
	"flag"
	"net/http"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/handler/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	flag.Parse()
	hub := ws.NewHub()
	go hub.Run()

	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "hello websocket",
		})
	})

	router.GET("/ws", wsHandler(hub))

	router.Run()
}

func wsHandler(hub *ws.Hub) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ws.ServerWs(hub, ctx.Writer, ctx.Request)
	}
}
