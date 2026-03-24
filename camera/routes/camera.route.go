package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CameraChannel1(c*gin.Context) {
   c.JSON(http.StatusOK, gin.H{
    "message": "Camera channel 1",
})
}

func CameraChannel2(c*gin.Context) {
   c.JSON(http.StatusOK, gin.H{
    "message": "Camera channel 2",
})
}

func CameraChannel3(c*gin.Context) {
   c.JSON(http.StatusOK, gin.H{
    "message": "Camera channel 3",
})
}

func CameraChannel4(c*gin.Context) {
   c.JSON(http.StatusOK, gin.H{
    "message": "Camera channel 4",
})
}
