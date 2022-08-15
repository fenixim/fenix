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

type FakeDatabaseError struct{}

func (f FakeDatabaseError) Error() string {
	return "FakeDatabaseError"
}

type InMemoryDatabase struct {
	ShouldErrorOnNext bool
	Messages          []*Message
	messagesLock      *sync.Mutex

	Users     map[string]*User
	usersLock *sync.Mutex

	Channels map[string]*Channel
	channelsLock *sync.Mutex
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		Users:        make(map[string]*User),
		usersLock:    &sync.Mutex{},
		messagesLock: &sync.Mutex{},
		Channels: make(map[string]*Channel),
		channelsLock: &sync.Mutex{},
	}
}

func (db *InMemoryDatabase) GetMessagesBetween(a, b, limit int64) ([]*Message, error) {
	db.messagesLock.Lock()
	defer db.messagesLock.Unlock()

	if db.ShouldErrorOnNext {
		return nil, FakeDatabaseError{}
	}

	partHistory := messages{}

	for _, m := range db.Messages {
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

	m.MessageID = primitive.NewObjectIDFromTimestamp(time.Unix(int64(len(db.Messages)+1), 0))
	db.Messages = append(db.Messages, m)
	return nil
}

func (db *InMemoryDatabase) InsertUser(u *User) error {
	db.usersLock.Lock()
	defer db.usersLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}

	u.UserID = primitive.NewObjectIDFromTimestamp(time.Unix(int64(len(db.Users)+1), 0))
	db.Users[u.UserID.Hex()] = u
	return nil
}

func (db *InMemoryDatabase) GetUser(req *User) error {
	db.usersLock.Lock()
	defer db.usersLock.Unlock()

	if db.ShouldErrorOnNext {
		return FakeDatabaseError{}
	}

	if req.UserID != primitive.NilObjectID {
		if user, ok := db.Users[req.UserID.Hex()]; ok {
			*req = *user
		} else {
			return DoesNotExist{}
		}
	} else if req.Username != "" {
		for _, u := range db.Users {
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

func (db *InMemoryDatabase) InsertChannel(channel *Channel) error {
	db.channelsLock.Lock()
	defer db.channelsLock.Unlock()
	channel.ChannelID = primitive.NewObjectID()
	db.Channels[channel.ChannelID.Hex()] = channel

	return nil
}

func (db *InMemoryDatabase) DeleteChannel(channel *Channel) error {
	db.channelsLock.Lock()
	defer db.channelsLock.Unlock()
	if _, ok := db.Channels[channel.ChannelID.Hex()]; !ok {
		return DoesNotExist{}
	}
	delete(db.Channels, channel.ChannelID.Hex())

	return nil
}
func (db *InMemoryDatabase) GetChannel(channel *Channel) error {
	db.channelsLock.Lock()
	defer db.channelsLock.Unlock()
	if c, ok := db.Channels[channel.ChannelID.Hex()]; ok {
		*channel = *c
	} else {
		return DoesNotExist{}
	}
	return nil
}