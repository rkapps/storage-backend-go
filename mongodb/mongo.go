package mongodb

import (
	"context"
	"reflect"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Create a new MongoClient with the connection string and dbname
func NewMongoClient(connstr string, dbname string) (*MongoClient, error) {

	ctx := context.Background()
	opts := options.Client().ApplyURI(connstr)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	db := client.Database(dbname)
	return &MongoClient{client, ctx, db}, nil
}

// Create a new MongoClient with the connection string and dbname
func NewMongoClientWithRegistry(connstr string, dbname string, reg *bson.Registry) (*MongoClient, error) {

	ctx := context.Background()

	opts := options.Client().
		ApplyURI(connstr).
		SetRegistry(reg)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	db := client.Database(dbname)
	return &MongoClient{client, ctx, db}, nil
}

func GetBsonRegistryForDecimal() *bson.Registry {

	// 1. Create a new registry (v2 directly uses NewRegistry)
	reg := bson.NewRegistry()
	// 2. Register your custom codec directly to the registry
	// DecimalCodec must implement EncodeValue and DecodeValue
	codec := &DecimalCodec{}
	reg.RegisterTypeEncoder(reflect.TypeOf(decimal.Decimal{}), codec)
	reg.RegisterTypeDecoder(reflect.TypeOf(decimal.Decimal{}), codec)
	return reg
}

// NewMongoRepository for the RepoModel
func NewMongoRepository[T RepoModel](client MongoClient) RepositoryInterface[T] {
	var model T
	collName := model.CollectionName()
	return &RepoCollection[T]{
		coll: client.DB.Collection(collName),
	}

}

// CreateSearchTokenField returns the token field
func CreateSearchTokenField(name string, values []string) SearchCriteriaTokenField {

	tokenField := SearchCriteriaTokenField{}
	tokenField.Name = name
	tokenField.Values = values
	return tokenField
}

// CreateSearchRangeField returns the range field
func CreateSearchRangeField(name string, key string, value float64) SearchCriteriaRangeField {

	rangeField := SearchCriteriaRangeField{}
	rangeField.Name = name
	rangeField.Key = key
	rangeField.Value = value
	return rangeField
}

// CreateSortField returns the sort field
func CreateSortField(name string, value int) SearchCriteriaSortField {
	sortField := SearchCriteriaSortField{}
	sortField.Name = name
	sortField.Value = value
	return sortField
}
