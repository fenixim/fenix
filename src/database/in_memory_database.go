package database

import (
	"fenix/src/utils"
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

type FakeDatabaseError struct{}

func (f FakeDatabaseError) Error() string {
	return "FakeDatabaseError"
}

type InMemoryDatabase struct {
	ShouldErrorOnNext bool
	messages          []*Message
	messagesLock      *sync.Mutex

	users     map[string]*User
	usersLock *sync.Mutex

	yodels     map[string]*Yodel
	yodelsLock *sync.Mutex
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		users:        make(map[string]*User),
		usersLock:    &sync.Mutex{},
		messagesLock: &sync.Mutex{},
		yodels:       make(map[string]*Yodel),
		yodelsLock:   &sync.Mutex{},
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

	m.MessageID = primitive.NewObjectIDFromTimestamp(time.Unix(int64(len(db.messages)+1), 0))
	db.messages = append(db.messages, m)
	return nil
}

func (db *InMemoryDatabase) InsertUser(u *User) error {
	db.usersLock.Lock()
	defer db.usersLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}

	u.UserID = primitive.NewObjectIDFromTimestamp(time.Unix(int64(len(db.users)+1), 0))
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
		utils.ErrorLogger.Println("GetUser needs fields in User!")
		return DatabaseError{}
	}

	return nil
}

func (db *InMemoryDatabase) InsertYodel(y *Yodel) error {
	db.yodelsLock.Lock()
	defer db.yodelsLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}

	y.YodelID = primitive.NewObjectIDFromTimestamp(time.Unix(int64(len(db.users)+1), 0))
	db.yodels[y.YodelID.Hex()] = y
	return nil
}

func (db *InMemoryDatabase) GetYodel(y *Yodel) error {
	db.yodelsLock.Lock()
	defer db.yodelsLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}
	yodel, ok := db.yodels[y.YodelID.Hex()]
	if !ok {
		return DoesNotExist{}
	}
	*y = *yodel
	return nil
}

func (db *InMemoryDatabase) ClearDB() error {
	db.messagesLock.Lock()
	db.messages = []*Message{}
	db.messagesLock.Unlock()

	db.usersLock.Lock()
	db.users = make(map[string]*User)
	db.usersLock.Unlock()

	db.yodelsLock.Lock()
	db.yodels = make(map[string]*Yodel)
	db.yodelsLock.Unlock()

	return nil
}
