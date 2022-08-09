package database

import (
	"log"
	"sort"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DoesNotExist struct {}

func (d *DoesNotExist) Error() string {
	return "Does Not Exist!"
}


type messages struct {
	M []Message
}

func (m messages) Less(i, j int) bool {
	return m.M[i].Timestamp < m.M[j].Timestamp
}

func (m messages) Swap(i, j int) {
	a := m.M[i]
	b := m.M[j]

	m.M[j] = a
	m.M[i] = b
}

func (m messages) Len() int {
	return len(m.M)
}


type StubDatabase struct {
	UsersById *sync.Map
	Messages *sync.Map
	UsersByUsername *sync.Map
}

func (s *StubDatabase) InsertMessage(m *Message) (error) {
	m.MessageID = primitive.NewObjectIDFromTimestamp(time.Unix(m.Timestamp, 0))
	s.Messages.Store(m.MessageID.Hex(), m)
	return nil
}

func (s *StubDatabase) GetMessagesBetween(a, b, limit int64) ([]Message, error) {
	msgs := make([]Message, 0)

	s.Messages.Range(func(key, value interface{}) bool {
		message := value.(*Message)
		if message.Timestamp >= a || message.Timestamp <= b {
			msgs = append(msgs, *message)
		}
		
		return true
	})
	m := messages{
		M: msgs,
	}

	sort.Stable(m)
	if int64(len(m.M)) >= limit {
		return m.M[:limit], nil
	}
	return m.M, nil
}

func (s *StubDatabase) GetMessage(m *Message) (error) {
	message, _ := s.Messages.Load(m.MessageID.Hex())
	typesMessage := (message).(*Message)
	*m = *typesMessage
	return nil
}

func (s *StubDatabase) DeleteMessage(m *Message) error {
	s.Messages.Delete(m.MessageID.Hex())
	return nil
}

func (s *StubDatabase) InsertUser(u *User) (error) {
	u.UserID = primitive.NewObjectIDFromTimestamp(time.Now())
	s.UsersById.Store(u.UserID.Hex(), u)
	s.UsersByUsername.Store(u.Username, u)
	return nil
}

func (s *StubDatabase) GetUser(u *User) (error) {
	var user interface{}
	var ok bool
	if u.UserID != primitive.NilObjectID {
		user, ok = s.UsersById.Load(u.UserID.Hex())
	} else if u.Username != "" {
		user, ok = s.UsersByUsername.Load(u.Username)
	} else {
		log.Panic("GetUser needs fields in User!")
	}
	if !ok {
		return &DoesNotExist{}
	}
	typedUser := (user).(*User)
	*u = *typedUser

	return nil
}

func (s *StubDatabase) DeleteUser(u *User) error {
	s.UsersById.Delete(u.UserID.Hex())
	s.UsersByUsername.Delete(u.Username)
	return nil
}
