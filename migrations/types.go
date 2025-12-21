package migrations

import (
	"strconv"
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

type MigrateFunc func(client *mongodb.MongoClient) error

func (m *Migration) Id() string {
	return strconv.Itoa(m.Version)
}

func (m *Migration) SetId() {
	m.ID = strconv.Itoa(m.Version)
}

func (m *Migration) CollectionName() string {
	return "migration"
}
