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
	partHistory := []Message{}
	start := int64(len(db.history))
	end := start - min(start, limit)
	for i := start - 1; i > end - 1; i-- {
		partHistory = append(partHistory, db.history[i])
	}
	return partHistory
}

func (db *InMemoryDatabase) InsertMessage(m *Message) {
	db.history = append(db.history, *m)
}
