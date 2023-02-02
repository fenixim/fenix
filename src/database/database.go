package database

import (
	"context"
	"fenix/src/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseError struct{}

func (e DatabaseError) Error() string {
	return "DatabaseError"
}

type Database interface {
	InsertMessage(*Message) error
	GetMessagesBetween(int64, int64, int64) ([]*Message, error)

	InsertUser(*User) error
	GetUser(*User) error

	InsertYodel(*Yodel) error
	GetYodel(*Yodel) error

	ClearDB() error
}

type MongoDatabase struct {
	mongo    *mongo.Client
	database string
}

func (db *MongoDatabase) getDatabase() *mongo.Database {
	mongoDB := db.mongo.Database(db.database)

	if mongoDB == nil {
		utils.ErrorLogger.Panicf("Must configure mongodb to have a %v database", db.database)
	}

	return mongoDB
}

func (db *MongoDatabase) makeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*5)
}

func (db *MongoDatabase) InsertYodel(y *Yodel) error {
	coll := db.getDatabase().Collection("yodels")

	ctx, cancel := db.makeContext()
	defer cancel()

	res, err := coll.InsertOne(ctx, y)

	y.YodelID = res.InsertedID.(primitive.ObjectID)
	return err
}

func (db *MongoDatabase) GetYodel(y *Yodel) error {
	coll := db.getDatabase().Collection("yodels")

	ctx, cancel := db.makeContext()
	defer cancel()
	q := bson.D{{
		"_id", bson.D{{
			"$eq", y.YodelID,
		}},
	}}

	res := coll.FindOne(ctx, q)

	err := res.Decode(y)
	return err
}

func (db *MongoDatabase) InsertMessage(m *Message) error {
	coll := db.getDatabase().Collection("messages")

	ctx, cancel := db.makeContext()
	defer cancel()

	res, err := coll.InsertOne(ctx, m)

	m.MessageID = res.InsertedID.(primitive.ObjectID)
	return err
}

func (db *MongoDatabase) GetMessagesBetween(a int64, b int64, limit int64) ([]*Message, error) {
	coll := db.getDatabase().Collection("messages")

	q := bson.D{{
		"$and",
		bson.A{
			bson.D{{"timestamp", bson.D{{"$gte", a}}}},
			bson.D{{"timestamp", bson.D{{"$lte", b}}}},
		},
	}}
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(limit)

	ctx, cancel := db.makeContext()
	defer cancel()

	cur, err := coll.Find(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	var res []*Message

	err = cur.All(context.Background(), &res)
	return res, err
}

func (db *MongoDatabase) InsertUser(u *User) error {
	coll := db.getDatabase().Collection("users")

	ctx, cancel := db.makeContext()
	defer cancel()

	res, err := coll.InsertOne(ctx, u)
	if err != nil {
		return err
	}

	u.UserID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (db *MongoDatabase) GetUser(u *User) error {
	coll := db.getDatabase().Collection("users")
	var q bson.D
	if u.UserID != primitive.NilObjectID {
		q = bson.D{{
			"_id", bson.D{{
				"$eq", u.UserID,
			}},
		}}
	} else if u.Username != "" {
		q = bson.D{{
			"username", bson.D{{
				"$eq", u.Username,
			}},
		}}
	} else {
		utils.ErrorLogger.Println("GetUser needs fields in User!")
		return DatabaseError{}
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

func (db *MongoDatabase) ClearDB() error {
	ctx, cancel := db.makeContext()
	defer cancel()

	return db.getDatabase().Drop(ctx)
}

func NewMongoDatabase(mongo_addr string, database string) *MongoDatabase {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(mongo_addr).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		utils.ErrorLogger.Panicf("Error connecting to mongoDB: %v", err)
	}

	db := MongoDatabase{
		mongo:    c,
		database: database,
	}
	return &db
}
