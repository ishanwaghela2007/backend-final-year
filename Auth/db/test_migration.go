package db

import (
    "fmt"
    "log"
    "time"
    "github.com/gocql/gocql"
    "golang.org/x/crypto/bcrypt"
)



// CreateUserTable creates the users table if it doesn't exist
func CreateTestUserTable() {
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
func BootstrapTestAdmin() {
    var id gocql.UUID
    err := Session.Query(`SELECT id FROM users WHERE email = ? LIMIT 1`,
        "admin@divyapackingtest.com").Scan(&id)

    if err == gocql.ErrNotFound {
        // ✅ Only create if admin truly doesn't exist
        adminID := gocql.TimeUUID()

        password := "admin123"
        hashed, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if hashErr != nil {
            log.Fatal("❌ Failed to hash password:", hashErr)
        }

        insertErr := Session.Query(`
            INSERT INTO users (id, name, email, password, role, isverified, isloggedin, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
            adminID,
            "Default Admin",
            "admin@divyapackingtest.com",
            string(hashed),
            "admin",
            true,
            false,
            time.Now(),
            time.Now(),
        ).Exec()

        if insertErr != nil {
            log.Fatal("❌ Failed to insert default admin:", insertErr)
        }

        fmt.Println("✅ Default admin created: email=admin@divyapackingtest.com password=admin123")
    } else if err != nil {
        log.Fatal("❌ Failed to query admin:", err)
    } else {
        fmt.Println("ℹ️ Admin already exists")
    }
}

