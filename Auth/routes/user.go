package routes

import (
	"Auth/db"
	"Auth/models"
	"fmt"
	"net/http"
	"time"
	"Auth/internal/kafka"
	"Auth/internal/emailjob"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *gin.Context) {
	var user models.User

	// Validate JSON input
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🔍 Check if email already exists
	var existingEmail string
	checkQuery := `SELECT email FROM users WHERE email = ? LIMIT 1`
	err := db.Session.Query(checkQuery, user.Email).Scan(&existingEmail)
	if err == nil {
		// Means record found
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists with this email"})
		return
	}

	// Hash password
	hashed, hashErr := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.ID = gocql.TimeUUID()
	user.CreatedAt = time.Now()
	user.IsLoggedIn = false

	// Insert new user
	insertQuery := `INSERT INTO users (id, name, email, password, role, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	if err := db.Session.Query(insertQuery,
		user.ID, user.Name, user.Email, string(hashed), user.Role, user.CreatedAt).Exec(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	emailJob := emailjob.EmailJob{
    To:       user.Email,
    Subject:  "Welcome to Our Platform",
    Body:     fmt.Sprintf("Hello %s, welcome aboard!", user.Name),
	Type:"welcome",
}
    kafka.PublishEmailJob(emailJob)
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"id":      user.ID,
		"email":   user.Email,
	})
}

func DeleteUser(c *gin.Context) {
	email := c.Param("email") // get email from URL

	// Delete user by email (primary key)
	query := `DELETE FROM users WHERE email = ?`
	if err := db.Session.Query(query, email).Exec(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
