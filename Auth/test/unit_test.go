package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/api/v1/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	})

	admin := router.Group("/api/v1/admin")
	admin.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	{
		admin.POST("/users", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"message": "User created"})
		})
		admin.DELETE("/users/:email", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
		})
	}

	return router
}

func TestLoginEndpoint(t *testing.T) {
	router := setupRouter()
	body, _ := json.Marshal(map[string]string{"email": "admin@example.com", "password": "admin15"})
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestCreateUserEndpoint(t *testing.T) {
	router := setupRouter()
	body := []byte(`{"name":"Test User"}`)
	req, _ := http.NewRequest("POST", "/api/v1/admin/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusCreated, resp.Code)
}

