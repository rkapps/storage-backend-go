package migrations

import (
	"time"

	"github.com/rkapps/storage-backend-go/mongodb"
)

type Migration struct {
	ID          string `json:"id" bson:"id"`
	Version     int
	Description string
	Up          MigrateFunc `bson:"-"`
	Down        MigrateFunc `bson:"-"`
	Timestamp   *time.Time
}

type MigrateFunc func(client *mongodb.MongoDatabase) error

func (m *Migration) Id() string {
	return m.ID
}

func (m *Migration) CollectionName() string {
	return "migration"
}
