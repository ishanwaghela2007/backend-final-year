package db

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserTable creates the users table if it doesn't exist
func CreateUserTable() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		email TEXT PRIMARY KEY,
		id UUID,
		name TEXT,
		password TEXT,
		role TEXT,
		isverified BOOLEAN,
		isloggedin BOOLEAN,
		verified_at TIMESTAMP,
		created_at TIMESTAMP
	);`

	if err := Session.Query(query).Exec(); err != nil {
		log.Fatal("❌ Error creating users table: ", err)
	}
	fmt.Println("✅ Users table is ready")
}

// BootstrapAdmin creates a default admin with hashed password if it doesn't exist
func BootstrapAdmin() {
	var id gocql.UUID
	err := Session.Query(`SELECT id FROM users WHERE email = ? LIMIT 1`,
		"admin@divyapacking.com").Scan(&id)

	if err == gocql.ErrNotFound {
		// ✅ Only create if admin truly doesn't exist
		adminID := gocql.TimeUUID()

		password := "admin123"
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			log.Fatal("❌ Failed to hash password:", hashErr)
		}

		insertErr := Session.Query(`
			INSERT INTO users (id, name, email, password, role, isverified, isloggedin, verified_at, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			adminID,
			"Default Admin",
			"admin@divyapacking.com",
			string(hashed),
			"admin",
			true,             // ✅ Admin is verified by default
			false,            // ✅ Admin is not logged in by default
			time.Now(),       // ✅ Verified timestamp
			time.Now(),
		).Exec()

		if insertErr != nil {
			log.Fatal("❌ Failed to insert default admin:", insertErr)
		}

		fmt.Println("✅ Default admin created: email=admin@divyapacking.com password=admin123")
	} else if err != nil {
		log.Fatal("❌ Failed to query admin:", err)
	} else {
		fmt.Println("ℹ️ Admin already exists")
	}
}
