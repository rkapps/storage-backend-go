package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SetupCollection creates the collection/indexes
func (repo *RepoCollection[T]) SetupCollection() error {
	var model T

	if ts := model.TimeSeriesOptionsBuilder(); ts != nil {

		opts := options.CreateCollection().SetTimeSeriesOptions(model.TimeSeriesOptionsBuilder())
		err := repo.coll.Database().CreateCollection(context.Background(), model.CollectionName(), opts)
		if err != nil {
			return err
		}

	}
	if imodels := model.IndexModels(); len(imodels) > 0 {
		_, err := repo.coll.Indexes().CreateMany(context.Background(), imodels)
		return err
	}
	return nil
}

// DeleteByID finds record from the collection using the id
func (repo *RepoCollection[T]) DeleteByID(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	_, err := repo.coll.DeleteOne(ctx, filter)
	return err
}

// FindByID finds record from the collection using the id
func (repo *RepoCollection[T]) FindByID(ctx context.Context, id string) (T, error) {
	filter := bson.M{"_id": id}
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

// Insert inserts a single record into the collection
func (repo *RepoCollection[T]) InsertOne(ctx context.Context, model RepoModel) error {
	_, err := repo.coll.InsertOne(ctx, model, nil)
	return err
}

// UpdateOne update a single record into the collection based on the id.
func (repo *RepoCollection[T]) UpdateOne(ctx context.Context, model RepoModel) error {
	update := bson.M{"$set": model}
	_, err := repo.coll.UpdateByID(ctx, model.Id(), update, nil)
	return err
}

// UpdateMany updates multiple records into the collection based on ids in the updateData
func (repo *RepoCollection[T]) UpdateManyByIds(ctx context.Context, ids []string, updateData interface{}) error {

	// The filter uses the $in operator to match any of the provided IDs
	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}
	_, err := repo.coll.UpdateMany(ctx, filter, updateData)
	return err
}
