package database

import (
	"errors"
	"github.com/astridalia/tinyrpg/models"
	"github.com/gocql/gocql"
)

var ErrImageNotFound = errors.New("Image not found")

type MyCassandraClient struct {
	session *gocql.Session
}

func InitCassandra() *MyCassandraClient {
	cluster := gocql.NewCluster("127.0.0.1") // Replace with your Cassandra cluster address

	// Create a temporary keyspace (not 'system') for this initialization
	cluster.Keyspace = "images"
	cluster.Consistency = gocql.Quorum // Choose the consistency level you need

	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}

	// Create the 'images' keyspace and table if they don't exist
	createKeyspaceAndTableIfNotExists(session)

	// Set the keyspace to 'images' in the cluster configuration
	cluster.Keyspace = "images"

	// Re-create the session with the 'images' keyspace set in the cluster configuration
	session, err = cluster.CreateSession()
	if err != nil {
		panic(err)
	}

	// Drop the temporary initialization keyspace
	session.Query("DROP KEYSPACE IF EXISTS temporary_init").Exec()

	MyCassandraClient := &MyCassandraClient{session: session}
	return MyCassandraClient
}

func createKeyspaceAndTableIfNotExists(session *gocql.Session) {
	// Create the 'images' keyspace if it doesn't exist
	query := session.Query(`
		CREATE KEYSPACE IF NOT EXISTS images WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }
	`)
	if err := query.Exec(); err != nil {
		panic(err)
	}

	// Create the 'images' table if it doesn't exist
	query = session.Query(`
		CREATE TABLE IF NOT EXISTS images (
			id UUID PRIMARY KEY,
			data TEXT
		)
	`)
	if err := query.Exec(); err != nil {
		panic(err)
	}
}

// InsertImage inserts an image into Cassandra.
func (c *MyCassandraClient) InsertImage(image *models.Image) error {
	query := c.session.Query(
		"INSERT INTO images (id, data) VALUES (?, ?)",
		image.ID,
		image.Data,
	)

	if err := query.Exec(); err != nil {
		return err
	}

	return nil
}

// GetImageFromCassandra retrieves an image from Cassandra by ID.
func (c *MyCassandraClient) GetImageFromCassandra(id string) (*models.Image, error) {
	var image models.Image

	query := c.session.Query(
		"SELECT id, data FROM images WHERE id = ?",
		id,
	)

	if err := query.Scan(&image.ID, &image.Data); err != nil {
		return nil, err
	}

	return &image, nil
}
