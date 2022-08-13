package database

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

type InMemoryDatabase struct {
	history []Message
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{}
}

func (db *InMemoryDatabase) GetMessagesBetween(_, _, limit int64) []Message {
	historySize := min(int64(len(db.history)), limit)
	return db.history[:historySize]
}

func (db *InMemoryDatabase) InsertMessage(m *Message) {
	db.history = append(db.history, *m)
}
