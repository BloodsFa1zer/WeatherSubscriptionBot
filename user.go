package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
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
		log.Warn().Err(err).Msg(" can`t connect to MongoDB")
		return nil
	}
	clientConn := ClientConnection{
		collection: client.Database("telegram").Collection("usersID"),
	}
	return &clientConn
}

func (cl *ClientConnection) findUser(field string, dataToFind any) *User {

	result := cl.collection.FindOne(context.TODO(), bson.M{field: dataToFind})

	// check for errors in the finding
	if result.Err() != nil {
		log.Warn().Err(result.Err()).Msg(" can`t find user")
	}

	// convert the cursor result to bson
	var user User
	// check for errors in the conversion
	if err := result.Decode(&user); err != mongo.ErrNoDocuments {
		log.Warn().Err(err).Msg(" no results to convert")
		return nil
	} else if err != nil {
		log.Warn().Err(err).Msg(" can`t convert results")
		return nil
	}

	return &user
}

func (cl *ClientConnection) createUser(user User) {
	userInfo := bson.D{{"UserID", user.UserID}, {"link", user.Link}, {"time", user.SendTime}}
	_, err := cl.collection.InsertOne(context.TODO(), userInfo)
	if err != nil {
		log.Warn().Err(err).Msg(" can`t insert user`s data into database")
		return
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
