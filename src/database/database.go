package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database interface {
	InsertMessage(*Message) (error)
	GetMessagesBetween(int64, int64, int64) ([]Message, error) 
	GetMessage(*Message) (error) 
	DeleteMessage(*Message) error

	InsertUser(*User) (error)
	GetUser(*User) (error)
	DeleteUser(*User) error
}


type MongoDatabase struct {
	mongo *mongo.Client
	database string
}

func (db *MongoDatabase) getDatabase() *mongo.Database {
	return db.mongo.Database(db.database)
}

func (db *MongoDatabase) makeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(),time.Second*5)
}

func (db *MongoDatabase) InsertMessage(m *Message) (error) {
	coll := db.getDatabase().Collection("messages")
	m.MessageID = primitive.NewObjectIDFromTimestamp(time.Unix(m.Timestamp, 0))

	ctx, cancel := db.makeContext()
	defer cancel()

	_, err := coll.InsertOne(ctx, m)
	return err
}

func (db *MongoDatabase) GetMessagesBetween(a int64, b int64, limit int64) ([]Message, error) {
	coll := db.getDatabase().Collection("messages")

	q := bson.D{{
		"$and",
		bson.A{
			bson.D{{"timestamp", bson.D{{"$gte", a}}}},
			bson.D{{"timestamp", bson.D{{"$lte", b}}}},
		},
	}}
	opts := options.Find().SetSort(bson.D{{"timestamp",1}}).SetLimit(limit)

	ctx, cancel := db.makeContext()
	defer cancel()

	cur, err := coll.Find(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	var res []Message

	err = cur.All(context.Background(), &res)
	return res, err
}

func (db *MongoDatabase) GetMessage(m *Message) (error) {
	coll := db.getDatabase().Collection("messages")
	q := bson.D{{
		"_id", bson.D{{
			"$eq", m.MessageID.Hex(),
		}},
	}}
	ctx, cancel := db.makeContext()
	defer cancel()

	res := coll.FindOne(ctx, q)

	err := res.Decode(m)
	return err
}

func (db *MongoDatabase) DeleteMessage(m *Message) error {
	coll := db.getDatabase().Collection("messages")
	q := bson.D{{
		"_id", bson.D{{
			"$eq", m.MessageID.Hex(),
		}},
	}}
	ctx, cancel := db.makeContext()
	defer cancel()

	_, err := coll.DeleteOne(ctx, q)

	return err
}

func (db *MongoDatabase) InsertUser(u *User) (error) {
	coll := db.getDatabase().Collection("messages")
	u.UserID = primitive.NewObjectIDFromTimestamp(time.Now())

	ctx, cancel := db.makeContext()
	defer cancel()

	_, err := coll.InsertOne(ctx, u)
	return err
}

func (db *MongoDatabase) GetUser(u *User) (error)  {
	coll := db.getDatabase().Collection("users")
	var q bson.D
	if u.UserID != primitive.NilObjectID {
		q = bson.D{{
			"_id", bson.D{{
				"$eq", u.UserID.Hex(),
			}},
		}}
	} else if u.Username != "" {
		q = bson.D{{
			"username", bson.D{{
				"$eq", u.Username,
			}},
		}}
	} else {
		log.Panic("GetUser needs fields in User!")
	}
	
	ctx, cancel := db.makeContext()
	defer cancel()

	res := coll.FindOne(ctx, q)

	err := res.Decode(u)
	return err
}

func (db *MongoDatabase) DeleteUser(u *User) error {
	coll := db.getDatabase().Collection("users")
	q := bson.D{{
		"_id", bson.D{{
			"$eq", u.UserID.Hex(),
		}},
	}}
	ctx, cancel := db.makeContext()
	defer cancel()

	_, err := coll.DeleteOne(ctx, q)

	return err
}

func NewMongoDatabase(mongo_addr string, database string) *MongoDatabase {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(mongo_addr).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, clientOptions)

	log.Fatalf("Error connecting to mongoDB: %v", err)
	db := MongoDatabase{
		mongo: c,
		database: database,
	}
	return &db
}
