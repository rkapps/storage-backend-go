package storage

import (
	"context"
	"log"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserModel struct {
	ID   string `bson:"_id"`
	Name string
}

func (u *UserModel) Id() string {
	return u.ID
}

func (u *UserModel) CollectionName() string {
	return "user"
}

func (u *UserModel) IndexModels() []mongo.IndexModel {

	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}}, // Type specified here
			Options: options.Index().SetUnique(false).SetName("idx_name"),
		},
	}
}

func (u *UserModel) TimeSeriesOptionsBuilder() *options.TimeSeriesOptionsBuilder {
	// return options.TimeSeries().SetTimeField("timestamp")
	return nil
}

func getClient() *MongoClient {

	client, err := NewMongoClient(os.Getenv("MONGO_ATLAS_CONN_STR"), "test")
	if err != nil {
		log.Fatalf("error connecting to client")
	}
	return client
}

func TestSetup(t *testing.T) {

	client := getClient()
	userRepo := NewMongoRepository[*UserModel](*client)
	err := userRepo.SetupCollection()
	if err != nil {
		t.Errorf("Expected no error: %v", err)
	}

}
func TestFindByID(t *testing.T) {

	client := getClient()
	userRepo := NewMongoRepository[*UserModel](*client)
	_, err := userRepo.FindByID(context.Background(), "123")
	if err != nil {
		t.Errorf("Expected no document: %v", err)
	} else {
		log.Println(err)
	}

}

func TestFind(t *testing.T) {

	client := getClient()
	userRepo := NewMongoRepository[*UserModel](*client)

	// The filter uses the $in operator to match any of the provided IDs
	// filter := bson.M{
	// 	"_id": bson.M{"$in": []string{""}},
	// }
	users, err := userRepo.Find(context.Background(), nil, bson.D{}, 0, 0)
	if err != nil {
		t.Errorf("Expected document: %v", err)
	} else {
		log.Println(len(users))
	}

}

func TestInsertOne(t *testing.T) {

	client := getClient()
	userRepo := NewMongoRepository[*UserModel](*client)

	user := &UserModel{}
	user.ID = "12345"
	user.Name = "ak"
	userRepo.InsertOne(context.Background(), user)
}

func TestUpdateOne(t *testing.T) {

	client := getClient()
	userRepo := NewMongoRepository[*UserModel](*client)

	user := &UserModel{}
	user.ID = "123"
	user.Name = "rk1"
	err := userRepo.UpdateOne(context.Background(), user)
	log.Println(err)
}
