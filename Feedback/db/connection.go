package db

import (
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
)

var Session *gocql.Session

func ConnectCassandra() {
	host := os.Getenv("CASSANDRA_HOST")
	keyspace := os.Getenv("CASSANDRA_KEYSPACE") // e.g., "auth" or a new one

	if host == "" {
		host = "127.0.0.1"
	}
	if keyspace == "" {
		keyspace = "chat"
	}

	// 1. Connect to system to create keyspace
	cluster := gocql.NewCluster(host)
	cluster.Consistency = gocql.Quorum
	
	// Create temporary session
	tempSession, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("❌ Failed to connect to Cassandra (System):", err)
	}
	
	query := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};`, keyspace)
	if err := tempSession.Query(query).Exec(); err != nil {
		log.Fatalf("❌ Failed to create keyspace '%s': %v", keyspace, err)
	}
	tempSession.Close()

	// 2. Connect to actual keyspace
	cluster.Keyspace = keyspace
	Session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal("❌ Failed to connect to Cassandra (Chat Keyspace):", err)
	}
	fmt.Printf("✅ Feedback Service: Connected to Keyspace '%s'\n", keyspace)
}

func Close() {
	if Session != nil {
		Session.Close()
		fmt.Println("🔌 Cassandra connection closed")
	}
}
