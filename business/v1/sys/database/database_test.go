package database_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"vtc/business/v1/sys/database"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

type User struct {
	Name    string `bson:"name"`
	Surname string `bson:"surname"`
	ID      string `bson:"_id"`
}

var (
	client *mongo.Database
	testID string
)

func TestMain(m *testing.M) {
	// open database client that will be share through all test
	var err error
	cfg := database.Config{
		Username:   "user",
		Password:   "password",
		Host:       "0.0.0.0",
		Port:       "20000",
		Database:   "thegoodseat_test",
		SSLEnabled: false,
	}

	client, err = database.NewClient(cfg)
	if err != nil {
		log.Fatalf("\t%s\t Test: \tShould be able to open a new client: %v", failure, err)
	}

	testID = uuid.NewString()

	//run test
	os.Exit(m.Run())
}

func Test_NewClient(t *testing.T) {
	t.Log("Given the need to be able to connect to the database")
	{
		cfg := database.Config{
			Username:   "user",
			Password:   "password",
			Host:       "0.0.0.0",
			Port:       "20000",
			Database:   "thegoodseat_test",
			SSLEnabled: false,
		}

		if _, err := database.NewClient(cfg); err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to open a new client: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to open a new client", success)
	}
}

func Test_Find(t *testing.T) {
	t.Log("Given the need to find many document")
	{
		res, err := database.Find[User](context.Background(), client, "test", bson.D{})
		if err != nil {
			t.Logf("\t%s\t Test: \tShould be able to find many document: %v", failure, err)
		}
		if len(res) < 1 {
			t.Logf("\t%s\t Test: \tShould be able to find many document: %v", failure, fmt.Errorf("failed to find more than 1 item with empty filter"))
		}
		t.Logf("\t%s\t Test: \tShould be able to find many document", success)
	}
}

func Test_FindOne(t *testing.T) {
	t.Log("Given the need to find one document")
	{
		var res User
		if err := database.FindOne[User](context.Background(), client, "test", bson.D{{"_id", testID}}, &res); err != nil {
			t.Logf("\t%s\t Test: \tShould be able to find one document: %v", failure, err)
		}
		if res.ID != testID {
			t.Logf("\t%s\t Test: \tShould be able to find one document: %v", failure, fmt.Errorf("test and res struct are not equal"))
		}
		t.Logf("\t%s\t Test: \tShould be able to find one document", success)
	}
}

func Test_InsertMany(t *testing.T) {
	t.Log("Given the need to insert many documents")
	{
		test := []User{
			{"Lorem", "IpSum", uuid.NewString()},
			{"Lorem", "IpSum", uuid.NewString()},
			{"Lorem", "IpSum", uuid.NewString()},
			{"Lorem", "IpSum", uuid.NewString()},
			{"Lorem", "IpSum", uuid.NewString()},
		}
		if err := database.InsertMany[User](context.Background(), client, "test", test); err != nil {
			t.Logf("\t%s\t Test: \tShould be able to insert many document: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to insert many document", success)
	}
}

func Test_InsertOne(t *testing.T) {
	test := User{"Samake", "Mahamadou", testID}

	t.Log("Given the need to insert document")
	{
		if err := database.InsertOne(context.Background(), client, "test", &test); err != nil {
			t.Logf("\t%s\t Test: \tShould be able to insert a document: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to insert a document", success)
	}
}

func Test_UpdateOne(t *testing.T) {
	t.Log("Given the need to update one document")
	{
		if err := database.UpdateOne[User](context.Background(), client, "test", testID, User{Name: "Bathie"}); err != nil {
			t.Logf("\t%s\t Test: \tShould be able to update one document: %v", failure, err)
		}
	}
}

func Test_DeleteOne(t *testing.T) {
	t.Log("Given the need to delete one document")
	{
		if err := database.DeleteOne(context.Background(), client, "test", testID); err != nil {
			t.Logf("\t%s\t Test: \tShould be able to delete one document: %v", failure, err)
		}
	}
}
