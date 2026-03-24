package db

import (
	"fmt"
	"log"
	"os"
	"github.com/gocql/gocql"
)

var Sessiontest *gocql.Session

func ConnecttestCassandra() {
	host := os.Getenv("CASSANDRATEST_HOST")
	keyspace := os.Getenv("CASSANDRATEST_KEYSPACE")

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

func ClosetestCassandra() {
	Sessiontest.Close()
	fmt.Println("🔌 Cassandra connection closed")
}
