package routes

import (
	httphandler "github.com/Harish-Naruto/Space-Striker-Server/internal/handler/http_handler"
	"github.com/gin-gonic/gin"
)

func RoomRoutes(router *gin.RouterGroup, h httphandler.Handler)  {
	router.GET("/room",h.RoomCreate)
}