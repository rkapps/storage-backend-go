package mongodb

import (
	"context"
	"time"

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
	CreateTimeSeriesCollection(ctx context.Context, timeField string, metaField string, dur time.Duration) error
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

// SearchCriteria holds fields used in AtlasSearch
type SearchCriteria struct {
	IndexName          string
	Query              string
	Limit              int
	AutoCompleteFields []string
	TokenFields        []SearchCriteriaTokenField
	RangeFields        []SearchCriteriaRangeField
	SortFields         []SearchCriteriaSortField
}

// SearchCriteriaTokenField has fields to use with "$in"
type SearchCriteriaTokenField struct {
	Name   string
	Values []string
}

// SearchCriteriaRangeField has fields to use with gt, gte, lt and lte
type SearchCriteriaRangeField struct {
	Name  string
	Key   string
	Value float64
}

// SearchCriteriaSortField has fields to use with sort name : -1 or name : -1
type SearchCriteriaSortField struct {
	Name  string
	Value int
}
