package mongodb

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/rkapps/storage-backend-go/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Use a standalone function
func GetMongoRepository[K comparable, M core.RepoModel[K]](db *MongoDatabase) core.Repository[K, M] {
	var model M
	collName := model.CollectionName()

	return &MongoRepository[K, M]{
		coll: db.collection(collName),
	}
}

// Aggregate returns aggregated documents based on pipeline
func (repo *MongoRepository[K, M]) Aggregate(ctx context.Context, pipeline interface{}, results any) error {

	cursor, err := repo.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	err = cursor.All(context.TODO(), results)
	return err
}

// BulkWrite writes multiple documents
func (repo *MongoRepository[K, M]) BulkWrite(ctx context.Context, ids []K, items []M) error {

	var operations []mongo.WriteModel
	for _, item := range items {
		operation := mongo.NewUpdateManyModel()
		operation.SetFilter(bson.M{"id": item.Id()})

		update := bson.M{"$set": item}
		operation.SetUpdate(update)
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}

	_, err := repo.coll.BulkWrite(ctx, operations)
	return err
}

// Count returns numbers of records in the collection
func (repo *MongoRepository[K, M]) Count(ctx context.Context) int64 {
	count, _ := repo.coll.CountDocuments(ctx, bson.D{}, nil)
	return count
}

// CreateIndexes creates the index for the collection.
func (repo *MongoRepository[K, M]) CreateIndexes(ctx context.Context, models []mongo.IndexModel) error {
	_, err := repo.coll.Indexes().CreateMany(ctx, models)
	return err
}

// CreateSearchIndexes creates the index for the collection.
func (repo *MongoRepository[K, M]) CreateSearchIndexes(ctx context.Context, models []mongo.SearchIndexModel) error {
	_, err := repo.coll.SearchIndexes().CreateMany(ctx, models, nil)
	return err
}

// CreateTimeSeriesCollection create a time series collection and secondary indexes
func (repo *MongoRepository[K, M]) CreateTimeSeriesCollection(ctx context.Context, timeField string, metaField string, dur time.Duration) error {

	//Create the collection
	tsOpts := options.TimeSeries()
	tsOpts.SetTimeField(timeField)
	tsOpts.SetMetaField(metaField)

	tsOpts.SetBucketMaxSpan(dur)
	tsOpts.SetBucketRounding(dur)

	cOpts := options.CreateCollection().SetTimeSeriesOptions(tsOpts)
	return repo.coll.Database().CreateCollection(ctx, repo.coll.Name(), cOpts)
}

// DeleteByID finds record from the collection using the id
func (repo *MongoRepository[K, M]) DeleteByID(ctx context.Context, id K) error {
	filter := bson.M{"id": id}
	_, err := repo.coll.DeleteOne(ctx, filter)
	return err
}

// DeleteMany delete multiple records in the collection
func (repo *MongoRepository[K, M]) DeleteMany(ctx context.Context, ids []K) error {
	var filter = bson.M{}
	if len(ids) > 0 {
		filter = bson.M{
			"id": bson.M{"$in": ids},
		}
	}
	_, err := repo.coll.DeleteMany(ctx, filter)
	return err
}

// FindByID finds record from the collection using the id
func (repo *MongoRepository[K, M]) FindByID(ctx context.Context, id K) (M, error) {
	filter := bson.M{"id": id}
	result := repo.coll.FindOne(ctx, filter)

	var model M
	if err := result.Decode(&model); err != nil {
		return model, err
	}
	return model, nil
}

// Find finds record from the collection by filter
func (repo *MongoRepository[K, M]) Find(ctx context.Context, filter any, sort bson.D, limit int64, skip int64) ([]M, error) {

	var models []M
	if sort == nil {
		sort = bson.D{}
	}
	opts := options.Find().SetSkip(skip).SetSort(sort).SetLimit(limit)
	result, err := repo.coll.Find(ctx, filter, opts)
	if err != nil {
		return models, err
	}

	if err := result.All(ctx, &models); err != nil {
		return models, err
	}
	return models, nil
}

// InsertOne inserts a single record into the collection
func (repo *MongoRepository[K, M]) InsertOne(ctx context.Context, item M) error {
	log.Println(item)
	_, err := repo.coll.InsertOne(ctx, item, nil)

	return err
}

// InsertMany inserts multiple records in the collection
func (repo *MongoRepository[K, M]) InsertMany(ctx context.Context, items []M) error {
	_, err := repo.coll.InsertMany(ctx, items, nil)

	return err
}

// Search searches the collection using the searchindex
func (repo *MongoRepository[K, M]) Search(ctx context.Context, criteria core.SearchCriteria) ([]M, error) {
	var results []M
	compound := bson.D{}

	if len(criteria.Query) > 0 {
		shouldClauses := bson.A{}
		for _, field := range criteria.AutoCompleteFields {
			shouldClauses = append(shouldClauses, bson.D{
				{Key: "autocomplete", Value: bson.D{
					{Key: "query", Value: criteria.Query},
					{Key: "path", Value: field},
				}},
			})
		}
		compound = append(compound, bson.E{Key: "should", Value: shouldClauses})
		compound = append(compound, bson.E{Key: "minimumShouldMatch", Value: 1})
	}

	//Add token fields for text fields and range fields for numerics
	if len(criteria.TokenFields) > 0 || len(criteria.RangeFields) > 0 || len(criteria.BooleanFields) > 0 {

		filterCause := bson.A{}
		if len(criteria.TokenFields) > 0 {
			for _, tokenField := range criteria.TokenFields {
				filterCause = append(filterCause, bson.D{
					{Key: "in", Value: bson.D{
						{Key: "path", Value: tokenField.Name},
						{Key: "value", Value: tokenField.Values},
					}},
				})
			}
		}
		if len(criteria.RangeFields) > 0 {
			for _, rangeField := range criteria.RangeFields {
				filterCause = append(filterCause, bson.D{
					{Key: "range", Value: bson.D{
						{Key: "path", Value: rangeField.Name},
						{Key: rangeField.Key, Value: rangeField.Value},
					}},
				})

			}
		}
		if len(criteria.BooleanFields) > 0 {
			for _, booleanField := range criteria.BooleanFields {
				filterCause = append(filterCause, bson.D{
					{Key: "equals", Value: bson.D{
						{Key: "path", Value: booleanField},
						{Key: "value", Value: true},
					}},
				})

			}
		}

		compound = append(compound, bson.E{Key: "filter", Value: filterCause})
	}

	searchValue := bson.D{}
	searchValue = append(searchValue, bson.E{Key: "index", Value: criteria.IndexName})
	searchValue = append(searchValue, bson.E{Key: "compound", Value: compound})

	//Add sortFields
	if len(criteria.SortFields) > 0 {
		sortValue := bson.D{}
		for _, field := range criteria.SortFields {
			sortValue = append(sortValue, bson.E{Key: field.Name, Value: field.Value})
		}
		sortStage := bson.E{Key: "sort", Value: sortValue}
		searchValue = append(searchValue, sortStage)
	}

	//Final searchStage
	searchStage := bson.D{
		{Key: "$search", Value: searchValue},
	}

	var pipeline []bson.D

	limitStage := bson.D{}
	if criteria.Limit > 0 {
		limitStage = append(limitStage, bson.E{Key: "$limit", Value: criteria.Limit})
		pipeline = mongo.Pipeline{searchStage, limitStage}
	} else {
		pipeline = mongo.Pipeline{searchStage}
	}

	slog.Debug("Search", "searchStage", fmt.Sprintf("%s", searchStage))
	cursor, err := repo.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return results, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return results, err
	}
	return results, nil
}

// UpdateOne update a single record into the collection based on the id.
func (repo *MongoRepository[K, M]) UpdateOne(ctx context.Context, item M) error {
	update := bson.M{"$set": item}
	_, err := repo.coll.UpdateByID(ctx, item.Id(), update, nil)
	return err
}

// UpdateMany updates multiple records into the collection based on ids in the update
func (repo *MongoRepository[K, M]) UpdateMany(ctx context.Context, ids []K, set bson.M) error {

	// The filter uses the $in operator to match any of the provided IDs
	filter := bson.M{
		"id": bson.M{"$in": ids},
	}
	update := bson.M{"$set": set}
	_, err := repo.coll.UpdateMany(ctx, filter, update)
	return err
}
