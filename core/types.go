package core

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RepoModel interface
type RepoModel[K any] interface {
	Id() K
	CollectionName() string
}

// Repository
type Repository[K any, M RepoModel[K]] interface {
	Aggregate(ctx context.Context, pipeline interface{}, results any) error
	BulkWrite(ctx context.Context, ids []K, items []M) error
	CreateIndexes(ctx context.Context, indexModels []mongo.IndexModel) error
	CreateSearchIndexes(ctx context.Context, searchIndexModels []mongo.SearchIndexModel) error
	CreateTimeSeriesCollection(ctx context.Context, timeField string, metaField string, dur time.Duration) error
	Count(ctx context.Context) int64
	DeleteByID(ctx context.Context, id K) error
	DeleteMany(ctx context.Context, ids []K) error
	FindByID(ctx context.Context, id K) (M, error)
	Find(ctx context.Context, filter any, sort bson.D, limit int64, skip int64) ([]M, error)
	InsertOne(ctx context.Context, item M) error
	InsertMany(ctx context.Context, items []M) error
	Search(ctx context.Context, criteria SearchCriteria) ([]M, error)
	UpdateOne(ctx context.Context, item M) error
	UpdateMany(ctx context.Context, ids []K, set bson.M) error
}

// SearchCriteria holds fields used in AtlasSearch
type SearchCriteria struct {
	IndexName          string
	Query              string
	Limit              int
	BooleanFields      []string
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

func (c SearchCriteria) AddSearchRangeField(name string, key string, value float64) {
	rangeField := SearchCriteriaRangeField{}
	rangeField.Name = name
	rangeField.Key = key
	rangeField.Value = value
	c.RangeFields = append(c.RangeFields, rangeField)
}

func (c SearchCriteria) AddSearchTokenField(name string, values []string) {
	tokenField := SearchCriteriaTokenField{}
	tokenField.Name = name
	tokenField.Values = values
	c.TokenFields = append(c.TokenFields, tokenField)
}

func (c SearchCriteria) AddBooleanField(name string) {
	c.BooleanFields = append(c.BooleanFields, name)
}

func (c SearchCriteria) AddSortField(name string, value int) {
	sortField := SearchCriteriaSortField{}
	sortField.Name = name
	sortField.Value = value
	c.SortFields = append(c.SortFields, sortField)
}
