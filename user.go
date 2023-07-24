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

type User struct {
	IDs      primitive.ObjectID `bson:"_id,omitempty"`
	UserID   int64              `bson:"UserID,omitempty"`
	Link     string             `bson:"link,omitempty"`
	SendTime string             `bson:"time, omitempty"`
}

type ClientConnection struct {
	collection *mongo.Collection
}

func newConnection(config Config) *ClientConnection {
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI_BD))
	if err != nil {
		panic(err)
		return nil
	}
	clientConn := ClientConnection{
		collection: client.Database("telegram").Collection("usersID"),
	}
	return &clientConn
}

func (cl *ClientConnection) findUser(field string, dataToFind any) (bool, primitive.ObjectID) {

	cursor, err := cl.collection.Find(context.TODO(), bson.M{field: dataToFind})
	// check for errors in the finding
	if err != nil {
		panic(err)
	}

	// convert the cursor result to bson
	var users []User
	// check for errors in the conversion
	if err = cursor.All(context.TODO(), &users); err != nil {
		panic(err)
	}

	// display the documents retrieved
	for _, u := range users {
		fmt.Println(u.IDs)
		return true, u.IDs

	}
	return false, [12]byte{}
}

func (cl *ClientConnection) createUser(user User) {
	userInfo := bson.D{{"UserID", user.UserID}, {"link", user.Link}, {"time", user.SendTime}}
	_, err := cl.collection.InsertOne(context.TODO(), userInfo)
	if err != nil {
		log.Panic().Err(err).Msg(" can`t insert user`s data into database")
	}
	log.Info().Msg("successfully insert user`s data")
}

func (cl *ClientConnection) updateUser(id *primitive.ObjectID, user User) {

	update := bson.M{"$set": bson.M{"UserID": user.UserID, "link": user.Link, "time": user.SendTime}}
	_, err := cl.collection.UpdateByID(context.Background(), id, update)
	if err != nil {
		panic(err)
	}

}
