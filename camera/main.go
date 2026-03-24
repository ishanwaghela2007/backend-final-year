package main

import (
	"camera/db"
	"camera/middleware"
	"camera/routes"
	"camera/utils"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/joho/godotenv"
)

func main() {
	//-----------env setup------------------
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("env loaded successfully")
    //----------redis setup----------------
	utils.ConnectRedis()
    //----------cassandra setup------------
	db.ConnectCassandra()
	//----------router setup---------------
	router := gin.Default()

	// -------------- CORS -----------------
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.NewRedisUserRateLimiter(utils.RDB, redis_rate.PerMinute(10)))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  http.StatusOK,
		})
	})

	cameraGroup := router.Group("api/v0/cctv")
	cameraGroup.Use(middleware.CameraAccess())
	{
		cameraGroup.GET("/stream/channel1", routes.CameraChannel1)
		cameraGroup.GET("/stream/channel2", routes.CameraChannel2)
		cameraGroup.GET("/stream/channel3", routes.CameraChannel3)
		cameraGroup.GET("/stream/channel4", routes.CameraChannel4)
	}
	port := os.Getenv("CAMERA_PORT")
	if port == "" {
		port = "3001"
	}
	router.Run(":" + port)
}