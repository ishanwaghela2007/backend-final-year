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
	keyspace := os.Getenv("CASSANDRA_KEYSPACE")

	cluster := gocql.NewCluster(host)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	var err error
	Session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal("❌ Failed to connect to Cassandra:", err)
	}
	fmt.Println("✅ Connected to Cassandra")
}

func Close() {
	Session.Close()
	fmt.Println("🔌 Cassandra connection closed")
}
