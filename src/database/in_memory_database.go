package database

import (
	"sort"
)

type InMemoryDatabase struct {
	history []*Message
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{}
}

func (db *InMemoryDatabase) GetMessagesBetween(a, b, limit int64) ([]*Message, error) {
	partHistory := messages{}

	for _, m := range db.history {
		if m.Timestamp > a && m.Timestamp <= b {
			partHistory.M = append(partHistory.M, m)
		}
	}

	sort.Sort(partHistory)
	if int64(len(partHistory.M)) >= limit {
		return partHistory.M[:limit], nil
	}
	return partHistory.M, nil
}

func (db *InMemoryDatabase) InsertMessage(m *Message) {
	db.history = append(db.history, m)
}
