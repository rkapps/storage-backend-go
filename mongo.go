package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Create a new MongoClient with the connection string and dbname
func NewMongoClient(connstr string, dbname string) (*MongoClient, error) {

	ctx := context.Background()
	client, err := mongo.Connect(options.Client().ApplyURI(connstr))
	if err != nil {
		return nil, err
	}
	db := client.Database(dbname)
	return &MongoClient{client, ctx, db}, nil
}

// NewMongoRepository for the RepoModel
func NewMongoRepository[T RepoModel](client MongoClient) RepositoryInterface[T] {
	var model T
	collName := model.CollectionName()
	return &RepoCollection[T]{
		coll: client.DB.Collection(collName),
	}

}
