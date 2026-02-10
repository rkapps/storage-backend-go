package mongodb

import (
	"github.com/rkapps/storage-backend-go/core"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoDatabase wraps the mongo client
type MongoDatabase struct {
	client *mongo.Client
	name   string
}

// MongoRepository wraps the mongo collection
type MongoRepository[K any, M core.RepoModel[K]] struct {
	coll *mongo.Collection
}
