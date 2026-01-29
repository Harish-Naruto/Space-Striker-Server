package httphandler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	
)

func (h Handler) RoomCreate(ctx *gin.Context)  {
	roomId ,err := h.HttpService.RoomGenerator()
	if err != nil {
		ctx.JSON(500,gin.H{
			"error":err,
		})
		return
	}
	ctx.JSON(http.StatusCreated,gin.H{
		"roomID":roomId,
	})
}