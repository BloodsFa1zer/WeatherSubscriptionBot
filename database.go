package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client mongo.Client

func MongoDBConnection(config Config) *mongo.Collection {
	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI_BD))
	if err != nil {
		panic(err)
		return nil
	}

	//defer client.Disconnect(ctx)

	usersCollection := client.Database("telegram").Collection("usersID")
	return usersCollection
}

func MongoDBFind(collection *mongo.Collection, field string, dataToFind any) (bool, primitive.ObjectID) {
	cursor, err := collection.Find(context.TODO(), bson.M{field: dataToFind})
	// check for errors in the finding
	if err != nil {
		panic(err)
		return false, [12]byte{}
	}

	// convert the cursor result to bson
	var users []User
	// check for errors in the conversion
	if err = cursor.All(context.TODO(), &users); err != nil {
		panic(err)
		return false, [12]byte{}
	}

	// display the documents retrieved
	fmt.Println("displaying all results in a collection")
	for _, u := range users {
		fmt.Println(u.IDs)
		return true, u.IDs
	}
	return false, [12]byte{}
}

func MongoDBWrite(collection *mongo.Collection, user User) {
	userInfo := bson.D{{"UserID", user.UserID}, {"link", user.Link}}
	_, err := collection.InsertOne(context.TODO(), userInfo)
	if err != nil {
		log.Panic().Err(err).Msg(" can`t insert user`s data into database")
	}
	log.Info().Msg("successfully insert user`s data")
}

func MongoDBUpdate(collection *mongo.Collection, id primitive.ObjectID, user User) {
	update := bson.D{
		{"$set", bson.D{{"UserID", user.UserID}}},
		{"$set", bson.D{{"link", user.Link}}},
	}
	_, err := collection.UpdateByID(context.TODO(), id, update)
	if err != nil {
		panic(err)
	}

}
