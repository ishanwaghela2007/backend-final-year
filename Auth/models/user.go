package models

import (
	"time"

	"github.com/gocql/gocql"
)

type User struct {
	ID         gocql.UUID `json:"id"`
	Name       string     `json:"name" binding:"required"`
	Email      string     `json:"email" binding:"required,email"`
	Password   string     `json:"password" binding:"required"`
	Role       string     `json:"role" binding:"required"`
	IsVerified bool       `json:"isverified"`
	IsLoggedIn bool       `json:"isloggedin"`
	VerifiedAt time.Time  `json:"verified_at"`
	CreatedAt  time.Time  `json:"created_at"`
}
