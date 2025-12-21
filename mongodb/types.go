package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoClient struct {
	client *mongo.Client
	ctx    context.Context
	DB     *mongo.Database
}

// RepoModel interface
type RepoModel interface {
	Id() string
	SetId()
	CollectionName() string
}

// RepoCollection
type RepoCollection[T RepoModel] struct {
	coll *mongo.Collection
}

// RepositoryInterface
type RepositoryInterface[T RepoModel] interface {
	BulkWrite(ctx context.Context, ids []string, items []T) error
	CreateIndexes(ctx context.Context, indexModels []mongo.IndexModel) error
	CreateSearchIndexes(ctx context.Context, searchIndexModels []mongo.SearchIndexModel) error
	Count(ctx context.Context) int64
	DeleteByID(ctx context.Context, id string) error
	DeleteMany(ctx context.Context, ids []string) error
	FindByID(ctx context.Context, id string) (T, error)
	Find(ctx context.Context, filter bson.M, sort bson.D, limit int64, skip int64) ([]T, error)
	InsertOne(ctx context.Context, item T) error
	InsertMany(ctx context.Context, items []T) error
	Search(ctx context.Context, criteria SearchCriteria) ([]T, error)
	UpdateOne(ctx context.Context, item T) error
	UpdateMany(ctx context.Context, ids []string, set bson.M) error
}

type SearchCriteria struct {
	Query              string
	AutoCompleteFields []string
	TokenFields        []SearchCriteriaTokenFields
	RangeFields        []SearchCriteriaRangeFields
}

type SearchCriteriaTokenFields struct {
	Name   string
	Values []string
}

type SearchCriteriaRangeFields struct {
	Name  string
	Key   string
	Value float64
}
