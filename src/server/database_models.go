package server

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func makeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second * 5)
}

type Message struct {
	MessageID primitive.ObjectID `bson:"_id"`
	Content string
	Timestamp int64
	Author string
}
func (m *Message) GetMessagesBetween(hub *ServerHub, a int64, b int64) (*[]Message, error) {
	c := hub.Database.Database(hub.MongoDatabase).Collection("messages")

	q := bson.D{{
		"$and",
		bson.A{
			bson.D{{"timestamp", bson.D{{"$gte", a}}}},
			bson.D{{"timestamp", bson.D{{"$lte", b}}}},
		},
	}}
	opts := options.Find().SetSort(bson.D{{"timestamp",1}})
	
	
	cur, err := c.Find(context.Background(), q, opts)
	if err != nil {
		return nil, err
	}
	var res []Message

	err = cur.All(context.Background(), &res)
	return &res, err
}
func (m *Message) InsertMessage(hub *ServerHub) {
	c := hub.Database.Database(hub.MongoDatabase).Collection("messages")
	ctx, cancel := makeContext()
	defer cancel()

	c.InsertOne(ctx, m)
}

type User struct {
	UserID primitive.ObjectID `bson:"_id"`
	Username string
	Password []byte
	Salt []byte
}

func (u *User) InsertUser(hub *ServerHub) (*mongo.InsertOneResult, error) {
	c := hub.Database.Database(hub.MongoDatabase).Collection("users")
	ctx, cancel := makeContext()
	defer cancel()

	return c.InsertOne(ctx, u)
}

func (u *User) FindUser(hub *ServerHub) error {
	c := hub.Database.Database(hub.MongoDatabase).Collection("users")
	ctx, cancel := makeContext()
	defer cancel()

	var q bson.D
	if u.UserID != primitive.NilObjectID {
		q = bson.D{{"_id", u.UserID}}
	} else if u.Username != "" {
		q = bson.D{{"username", u.Username}}
	} else {
		panic("no value set for finduser")
	}

	res := c.FindOne(ctx, q) 
	return res.Decode(u)
}