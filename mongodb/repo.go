package mongodb

import (
	"context"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Aggregate returns aggregated documents based on pipeline
func (repo *RepoCollection[T]) Aggregate(ctx context.Context, pipeline interface{}) ([]map[string]interface{}, error) {

	cursor, err := repo.coll.Aggregate(ctx, pipeline)
	var results []map[string]interface{}
	err = cursor.All(context.TODO(), &results)
	return results, err
}

// BulkWrite writes multiple documents
func (repo *RepoCollection[T]) BulkWrite(ctx context.Context, ids []string, items []T) error {

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
func (repo *RepoCollection[T]) Count(ctx context.Context) int64 {
	count, _ := repo.coll.CountDocuments(ctx, bson.D{}, nil)
	return count
}

// CreateIndexes creates the index for the collection.
func (repo *RepoCollection[T]) CreateIndexes(ctx context.Context, models []mongo.IndexModel) error {
	_, err := repo.coll.Indexes().CreateMany(ctx, models)
	return err
}

// CreateSearchIndexes creates the index for the collection.
func (repo *RepoCollection[T]) CreateSearchIndexes(ctx context.Context, models []mongo.SearchIndexModel) error {
	_, err := repo.coll.SearchIndexes().CreateMany(ctx, models, nil)
	return err
}

// CreateTimeSeriesCollection create a time series collection and secondary indexes
func (repo *RepoCollection[T]) CreateTimeSeriesCollection(ctx context.Context, timeField string, metaField string, dur time.Duration) error {

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
func (repo *RepoCollection[T]) DeleteByID(ctx context.Context, id string) error {
	filter := bson.M{"id": id}
	_, err := repo.coll.DeleteOne(ctx, filter)
	return err
}

// DeleteMany delete multiple records in the collection
func (repo *RepoCollection[T]) DeleteMany(ctx context.Context, ids []string) error {
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
func (repo *RepoCollection[T]) FindByID(ctx context.Context, id string) (T, error) {
	filter := bson.M{"id": id}
	result := repo.coll.FindOne(ctx, filter)

	var model T
	if err := result.Decode(&model); err != nil {
		return model, err
	}
	return model, nil
}

// Find finds record from the collection by filter
func (repo *RepoCollection[T]) Find(ctx context.Context, filter bson.M, sort bson.D, limit int64, skip int64) ([]T, error) {

	var models []T
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
func (repo *RepoCollection[T]) InsertOne(ctx context.Context, item T) error {
	log.Println(item)
	_, err := repo.coll.InsertOne(ctx, item, nil)

	return err
}

// InsertMany inserts multiple records in the collection
func (repo *RepoCollection[T]) InsertMany(ctx context.Context, items []T) error {
	_, err := repo.coll.InsertMany(ctx, items, nil)

	return err
}

// Search searches the collection using the searchindex
func (repo *RepoCollection[T]) Search(ctx context.Context, criteria SearchCriteria) ([]T, error) {
	var results []T
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
	if len(criteria.TokenFields) > 0 || len(criteria.RangeFields) > 0 {

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

	limitStage := bson.D{}
	if criteria.Limit > 0 {
		limitStage = append(limitStage, bson.E{Key: "$limit", Value: criteria.Limit})
	}

	//Final searchStage
	searchStage := bson.D{
		{Key: "$search", Value: searchValue},
	}

	slog.Debug("Search", "searchStage", searchStage.String())

	pipeline := mongo.Pipeline{searchStage, limitStage}
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
func (repo *RepoCollection[T]) UpdateOne(ctx context.Context, item T) error {
	update := bson.M{"$set": item}
	_, err := repo.coll.UpdateByID(ctx, item.Id(), update, nil)
	return err
}

// UpdateMany updates multiple records into the collection based on ids in the update
func (repo *RepoCollection[T]) UpdateMany(ctx context.Context, ids []string, set bson.M) error {

	// The filter uses the $in operator to match any of the provided IDs
	filter := bson.M{
		"id": bson.M{"$in": ids},
	}
	update := bson.M{"$set": set}
	_, err := repo.coll.UpdateMany(ctx, filter, update)
	return err
}
