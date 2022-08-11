package database

import (
	"crypto/sha512"

	"github.com/xdg-go/pbkdf2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	MessageID primitive.ObjectID `bson:"_id"`
	Content   string
	Timestamp int64
	Author    string
}

type User struct {
	UserID   primitive.ObjectID `bson:"_id"`
	Username string
	Password []byte
	Salt     []byte
}

func (u *User) HashPassword() {
	u.Password = pbkdf2.Key(u.Password, u.Salt, 100000, 32, sha512.New512_256)
}
