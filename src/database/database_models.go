package database

import (
	"crypto/sha512"
	"time"

	"github.com/xdg-go/pbkdf2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	MessageID primitive.ObjectID `bson:"_id,omitempty"`
	Content   string
	Timestamp int64
	Author    User
}

type User struct {
	UserID   primitive.ObjectID `bson:"_id,omitempty"`
	Username string
	Password []byte `json:"-"`
	Salt     []byte `json:"-"`
}

func (u *User) HashPassword() {
	u.Password = pbkdf2.Key(u.Password, u.Salt, 100000, 32, sha512.New512_256)
}

func NewMessage(user User, content string) *Message {
	m := Message{Author: user, Content: content, Timestamp: time.Now().UnixNano()}

	return &m
}
