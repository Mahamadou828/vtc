package database_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"vtc/business/v1/sys/database"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func TestDatabase(t *testing.T) {
	type Test struct {
		Name    string `bson:"name"`
		Surname string `bson:"surname"`
		ID      string `bson:"_id"`
	}

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

		client, err := database.NewClient(cfg)
		if err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to open a new client: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to open a new client", success)

		t.Log("Given the need to be able to interact with the database")
		{
			testID := uuid.NewString()
			test := Test{"Samake", "Mahamadou", testID}

			t.Log("Given the need to insert document")
			{
				if err := database.InsertOne(context.Background(), client, "test", test); err != nil {
					t.Logf("\t%s\t Test: \tShould be able to insert a document: %v", failure, err)
				}
				t.Logf("\t%s\t Test: \tShould be able to insert a document", success)
			}

			t.Log("Given the need to insert many documents")
			{
				test := []Test{
					{"Lorem", "IpSum", uuid.NewString()},
					{"Lorem", "IpSum", uuid.NewString()},
					{"Lorem", "IpSum", uuid.NewString()},
					{"Lorem", "IpSum", uuid.NewString()},
					{"Lorem", "IpSum", uuid.NewString()},
				}
				if err := database.InsertMany[Test](context.Background(), client, "test", test); err != nil {
					t.Logf("\t%s\t Test: \tShould be able to insert many document: %v", failure, err)
				}
				t.Logf("\t%s\t Test: \tShould be able to insert many document", success)
			}

			t.Log("Given the need to find many document")
			{
				res, err := database.Find[Test](context.Background(), client, "test", bson.D{})
				if err != nil {
					t.Logf("\t%s\t Test: \tShould be able to find many document: %v", failure, err)
				}
				if len(res) < 1 {
					t.Logf("\t%s\t Test: \tShould be able to find many document: %v", failure, fmt.Errorf("failed to find more than 1 item with empty filter"))
				}
				t.Logf("\t%s\t Test: \tShould be able to find many document", success)
			}

			t.Log("Given the need to find one document")
			{
				var res Test
				if err := database.FindOne[Test](context.Background(), client, "test", bson.D{{"_id", testID}}, &res); err != nil {
					t.Logf("\t%s\t Test: \tShould be able to find one document: %v", failure, err)
				}
				if res != test {
					t.Logf("\t%s\t Test: \tShould be able to find one document: %v", failure, fmt.Errorf("test and res struct are not equal"))
				}
				t.Logf("\t%s\t Test: \tShould be able to find one document", success)
			}

			t.Log("Given the need to update one document")
			{
				if err := database.UpdateOne[Test](context.Background(), client, "test", testID, Test{Name: "Bathie"}); err != nil {
					t.Logf("\t%s\t Test: \tShould be able to update one document: %v", failure, err)
				}
			}

			t.Log("Given the need to delete one document")
			{
				if err := database.DeleteOne(context.Background(), client, "test", testID); err != nil {
					t.Logf("\t%s\t Test: \tShould be able to delete one document: %v", failure, err)
				}
			}
		}
	}
}
