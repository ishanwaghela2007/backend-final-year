package main

import (
	"Auth/db"
	"Auth/internal/kafka"
	"Auth/middleware"
	"Auth/routes"
	"Auth/utils"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/joho/godotenv"
)

func main() {
	// ---------------- Load environment ----------------
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found — using system defaults")
	}

	port := getEnv("PORT", "8080")
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")
	setProxyEnv(httpProxy, httpsProxy)

	// ---------------- Database setup ----------------
	db.ConnectCassandra()
	defer db.Close()
	db.CreateUserTable()
	db.BootstrapAdmin()

	// ---------------- Redis setup ----------------
	utils.ConnectRedis()
	defer func() {
		if utils.RDB != nil {
			_ = utils.RDB.Close()
		}
	}()

	// ---------------- Kafka setup ----------------
	brokers := []string{"localhost:9092"}
	dlqTopic := "email_dlq"

	if err := kafka.InitProducer(brokers); err != nil {
		log.Fatalf("❌ Kafka producer init failed: %v", err)
	}
	defer kafka.CloseProducer()
	log.Println("✅ Kafka producer initialized")
	// Start consumer (email worker)
	go func() {
		ctx := context.Background()
		if err := kafka.StartEmailConsumer(ctx, brokers, kafka.Producer, dlqTopic); err != nil {
			log.Fatalf("❌ Kafka consumer failed: %v", err)
		}
	}()
	log.Println("✅ Email consumer worker started")

	// ---------------- Gin setup ----------------
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	rateLimit := redis_rate.PerMinute(10)
	router.Use(middleware.NewRedisUserRateLimiter(utils.RDB, rateLimit))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Auth microservice running",
		})
	})

	// Public routes
	api := router.Group("/api/v0")
	{
		api.POST("/login", routes.Login)
		api.GET("/oauth", routes.Oauthlogin)
	}

	// Protected routes
	protected := router.Group("/api/v0")
	protected.Use(middleware.AuthMiddleware(""))
	{
		protected.POST("/logout", routes.Logout)
		protected.GET("/me", routes.GetProfile)
	}

	// Admin routes
	admin := router.Group("/api/v0/admin")
	admin.Use(middleware.AuthMiddleware("admin"))
	{
		admin.POST("/users", routes.CreateUser)
		admin.DELETE("/users/:email", routes.DeleteUser)
	}

	// ---------------- Graceful Shutdown ----------------
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 Server running on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️ Shutting down server gracefully...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("❌ Server Shutdown Failed:%+v", err)
	}
	log.Println("✅ Server exited cleanly")
}

// ---------------- Helpers ----------------
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func setProxyEnv(httpProxy, httpsProxy string) {
	if httpProxy != "" {
		os.Setenv("HTTP_PROXY", httpProxy)
	}
	if httpsProxy != "" {
		os.Setenv("HTTPS_PROXY", httpsProxy)
	}
}
