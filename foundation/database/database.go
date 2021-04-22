package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrForbidden             = errors.New("attempted action is not allowed")
)

//DBinstance func
func DBinstance() *mongo.Client {

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client
}

//Client Database instance
var Client *mongo.Client = DBinstance()

//OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("drop").Collection(collectionName)

	return collection
}

// func StatusCheck(ctx context.Context, db *mongo.Client) error {
// 	// First check we can ping the database.
// 	var pingError error
// 	for attempts := 1; ; attempts++ {
// 		pingError = db.Ping(ctx, nil)
// 		if pingError == nil {
// 			break
// 		}
// 		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
// 		if ctx.Err() != nil {
// 			return ctx.Err()
// 		}
// 	}

// 	// Make sure we didn't timeout or be cancelled.
// 	if ctx.Err() != nil {
// 		return ctx.Err()
// 	}

// 	return nil
// }
