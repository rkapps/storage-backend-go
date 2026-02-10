package mongodb

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserModel struct {
	ID     string `bson:"id"`
	UserId string `bson:"user_id"`
	Name   string
}

func (u *UserModel) Id() string {
	return u.ID
}

func (u *UserModel) SetId() {
	u.ID = u.UserId
}

func (u *UserModel) CollectionName() string {
	return "user"
}

func getMongoDatabase() *MongoDatabase {

	database, err := NewMongoDatabase(os.Getenv("MONGO_ATLAS_CONN_STR"), "test")
	if err != nil {
		log.Fatalf("error connecting to client")
	}
	return database
}

func TestUserCollection(t *testing.T) {

	database := getMongoDatabase()
	userRepo := GetMongoRepository[string, *UserModel](database)
	ctx := context.Background()

	t.Run("clean", func(t *testing.T) {
		userRepo.DeleteMany(ctx, nil)
	})

	t.Run("insert", func(t *testing.T) {

		user := &UserModel{}
		user.UserId = "1"
		user.Name = "ak"
		user.SetId()
		userRepo.InsertOne(context.Background(), user)

		user1 := &UserModel{}
		user1.UserId = "2"
		user1.Name = "bk"
		user1.SetId()

		user2 := &UserModel{}
		user2.UserId = "3"
		user2.Name = "ck"
		user2.SetId()

		users := []*UserModel{}
		users = append(users, user1)
		users = append(users, user2)

		err := userRepo.InsertMany(ctx, users)
		if err != nil {
			t.Errorf("Error inserting many: %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {

		set := bson.M{"name": "dk"}
		err := userRepo.UpdateMany(ctx, []string{"3"}, set)
		if err != nil {
			t.Errorf("Error updating many: %v", err)
		}
		user, err := userRepo.FindByID(ctx, "3")
		if err != nil {
			t.Errorf("Error finding user: %v :%v", "3", err)
		}
		if strings.Compare(user.Name, "dk") != 0 {
			t.Errorf("Expecting user name to be '%v' %v", "dk", err)
		}
	})

	t.Run("find", func(t *testing.T) {

		sort := bson.D{{Key: "name", Value: -1}}
		users, _ := userRepo.Find(ctx, bson.D{}, sort, 0, 0)
		if len(users) == 0 {
			t.Error("Expecting users")
		} else {
			log.Printf("Users: %v", len(users))
			for _, user := range users {
				log.Println(user.Name)
			}
		}

	})
}
