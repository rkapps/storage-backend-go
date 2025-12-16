package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoClient struct {
	client *mongo.Client
	ctx    context.Context
	DB     *mongo.Database
}

// RepoModel interface
type RepoModel interface {
	Id() string
	CollectionName() string
	IndexModels() []mongo.IndexModel
	TimeSeriesOptionsBuilder() *options.TimeSeriesOptionsBuilder
}

// RepoCollection
type RepoCollection[T RepoModel] struct {
	coll *mongo.Collection
}

// RepositoryInterface
type RepositoryInterface[T RepoModel] interface {
	DeleteByID(ctx context.Context, id string) error
	SetupCollection() error
	FindByID(ctx context.Context, id string) (T, error)
	Find(ctx context.Context, filter bson.M, sort bson.D, limit int64, skip int64) ([]T, error)
	InsertOne(ctx context.Context, model RepoModel) error
	UpdateOne(ctx context.Context, model RepoModel) error
	UpdateManyByIds(ctx context.Context, ids []string, updateData interface{}) error
}
