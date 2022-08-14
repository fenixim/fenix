package database

import (
	"log"
	"sort"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DoesNotExist struct{}

func (d DoesNotExist) Error() string {
	return "Does Not Exist!"
}

type messages struct {
	M []*Message
}

func (m messages) Less(i, j int) bool {
	return m.M[i].Timestamp >= m.M[j].Timestamp
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

type FakeDatabaseError struct {}
func (f FakeDatabaseError) Error() string {
	return "FakeDatabaseError"
}

type InMemoryDatabase struct {
	ShouldErrorOnNext bool
	messages []*Message
	messagesLock *sync.Mutex

	users map[string]*User
	usersLock *sync.Mutex
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		users: make(map[string]*User),
		usersLock: &sync.Mutex{},
		messagesLock: &sync.Mutex{},
	}
}

func (db *InMemoryDatabase) GetMessagesBetween(a, b, limit int64) ([]*Message, error) {
	db.messagesLock.Lock()
	defer db.messagesLock.Unlock()

	if db.ShouldErrorOnNext {
		return nil, FakeDatabaseError{}
	}

	partHistory := messages{}

	for _, m := range db.messages {
		if m.Timestamp >= a && m.Timestamp <= b {
			partHistory.M = append(partHistory.M, m)
		}
	}

	sort.Sort(partHistory)
	if int64(len(partHistory.M)) >= limit {
		return partHistory.M[:limit], nil
	}
	return partHistory.M, nil
}

func (db *InMemoryDatabase) InsertMessage(m *Message) error {
	db.messagesLock.Lock()
	defer db.messagesLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}

	db.messages = append(db.messages, m)
	return nil
}

func (db *InMemoryDatabase) InsertUser(u *User) error {
	db.usersLock.Lock()
	defer db.usersLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}

	u.UserID = primitive.NewObjectIDFromTimestamp(time.Unix(int64(len(db.users) + 1), 0))
	db.users[u.UserID.Hex()] = u
	return nil
}

func (db *InMemoryDatabase) GetUser(req *User) error {
	db.usersLock.Lock()
	defer db.usersLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}

	if req.UserID != primitive.NilObjectID {
		if user, ok := db.users[req.UserID.Hex()]; ok {
			*req = *user
		} else {
			return DoesNotExist{}
		}
	} else if req.Username != "" {
		for _, u := range db.users {
			if u.Username == req.Username {
				*req = *u
				return nil
			}
		}
		return DoesNotExist{}
	} else {
		log.Panic("GetUser needs fields in User!")
	}

	return nil
}