package main

import (
	"flag"
	"net/http"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/handler/http_handler"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/handler/routes"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/infra"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/services"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/handler/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	flag.Parse()
	
	rdb := infra.CreateRedisClient("localhost:6379")
	hub := ws.NewHub(rdb)
	hs := services.CreateHttpService(rdb)
	go hub.Run()
	
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "hello websocket",
		})
	})

	v1 := router.Group("/api/v1")

	routes.RoomRoutes(v1,httphandler.Handler{
		HttpService: hs,
	})

	router.GET("/ws", wsHandler(hub))

	router.Run(":8080")
}

func wsHandler(hub *ws.Hub) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ws.ServerWs(hub, ctx.Writer, ctx.Request)
	}
}
