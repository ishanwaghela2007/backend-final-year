package db

import (
	"fmt"
	"log"
)

func CreateMessageTable() {
	// CLUSTERING ORDER BY created_at ASC allows retrieving messages in chronological order easily
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		channel_id TEXT,
		id UUID,
		sender_id TEXT,
		sender_role TEXT,
		content TEXT,
		created_at TIMESTAMP,
		PRIMARY KEY (channel_id, created_at)
	) WITH CLUSTERING ORDER BY (created_at ASC);`

	if err := Session.Query(query).Exec(); err != nil {
		log.Printf("❌ Error creating messages table: %v", err)
		return
	}
	fmt.Println("✅ Messages table is ready")
}
