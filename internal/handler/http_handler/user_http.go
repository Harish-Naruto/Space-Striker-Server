package httphandler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// add room validator to check if room is valid
func RoomRoutes(router *gin.RouterGroup) {
	router.GET("/room", func(ctx *gin.Context) {
		roomID, err := generateRoomID(5)
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "failed to generate RoomID",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"roomID": roomID,
		})

		return
	})
}

func generateRoomID(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
