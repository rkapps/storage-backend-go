package migrations

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sort"
	"strconv"
	"time"

	mongodb "github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var migrationsm map[string]map[int]*Migration

// Register registers migration versions
func Register(dbname string, version int, description string, up MigrateFunc, down MigrateFunc) {

	log.Printf("Reached register: %s", dbname)
	if exists := migrationsm[dbname][version]; exists != nil {
		panic(fmt.Sprintf("migration version '%d' already exists", version))
	}
	timestamp := time.Now()
	migration := &Migration{Version: version, Description: description, Up: up, Down: down, Timestamp: &timestamp}
	migration.ID = strconv.Itoa(migration.Version)

	if migrationsm == nil {
		migrationsm = make(map[string]map[int]*Migration)
	}
	if migrationsm[dbname] == nil {
		migrationsm[dbname] = make(map[int]*Migration)
	}

	migrationsm[dbname][version] = migration
	log.Println(migrationsm)
}

// RunMigrations runs all migrations
func RunMigrations(database *mongodb.MongoDatabase) error {

	migrations := getMigrations(database.Name())
	log.Printf("migrations for %s: %d", database.Name(), len(migrations))
	model := mongodb.GetMongoRepository[string, *Migration](database)
	cmigrations, err := model.Find(context.Background(), bson.M{}, bson.D{{Key: "version", Value: -1}}, 0, 0)
	if err != nil {
		return err
	}
	var cversion = 0
	if len(cmigrations) > 0 {
		cversion = cmigrations[0].Version
	}

	slog.Info(fmt.Sprintf("Migrations current Version: %d", cversion))
	for _, migration := range migrations {
		// log.Printf("ID: %s Version: %d Description: %s", migration.ID, migration.Version, migration.Description)
		slog.Info("Migration ID: "+migration.ID, "Version", migration.Version, "Description", migration.Description)
		if migration.Version <= cversion {
			continue
		}
		err := migration.Up(database)
		if err != nil {
			return fmt.Errorf("Error running migration %d:%s - %v", migration.Version, migration.Description, err)
		}
		err = model.InsertOne(context.Background(), migration)
		if err != nil {
			return fmt.Errorf("Error inserting migration record: %v", err)
		}

	}
	return nil
}

func getMigrations(dbname string) []*Migration {

	var migrations []*Migration
	for _, migration := range migrationsm[dbname] {
		migrations = append(migrations, migration)
	}
	sort.Slice(migrations, func(i int, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations
}
