package migrations

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	mongodb "github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var migrationsm map[int]*Migration

// Register registers migration versions
func Register(version int, description string, up MigrateFunc, down MigrateFunc) {

	if exists := migrationsm[version]; exists != nil {
		panic(fmt.Sprintf("migration version '%d' already exists", version))
	}
	timestamp := time.Now()
	migration := &Migration{Version: version, Description: description, Up: up, Down: down, Timestamp: &timestamp}
	migration.SetId()

	if migrationsm == nil {
		migrationsm = make(map[int]*Migration)
	}
	migrationsm[version] = migration
}

// RunMigrations runs all migrations
func RunMigrations(client *mongodb.MongoClient) error {

	migrations := getMigrations()
	model := mongodb.NewMongoRepository[*Migration](*client)
	cmigrations, err := model.Find(context.Background(), bson.M{}, bson.D{{Key: "version", Value: -1}}, 0, 0)
	if err != nil {
		return err
	}
	var cversion = 0
	if len(cmigrations) > 0 {
		cversion = cmigrations[0].Version
	}

	log.Printf("Current Version: %d", cversion)

	for _, migration := range migrations {
		log.Printf("ID: %s Version: %d Description: %s", migration.ID, migration.Version, migration.Description)
		if migration.Version <= cversion {
			continue
		}
		err := migration.Up(client)
		if err != nil {
			return fmt.Errorf("Error running migration %d:%s - %v", migration.Version, migration.Description, err)
		}
		// mmodel := &Migration{}
		// mmodel.Version = migration.Version
		// mmodel.SetId()
		// mmodel.Description = migration.Description
		// mmodel.Timestamp = migration.Timestamp
		err = model.InsertOne(context.Background(), migration)
		if err != nil {
			return fmt.Errorf("Error inserting migration record: %v", err)
		}

	}
	return nil
}

func getMigrations() []*Migration {

	var migrations []*Migration
	for _, migration := range migrationsm {
		migrations = append(migrations, migration)
	}
	sort.Slice(migrations, func(i int, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations
}
