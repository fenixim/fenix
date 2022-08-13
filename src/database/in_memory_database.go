package database

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

type InMemoryDatabase struct {
	size int64
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{}
}

func (db *InMemoryDatabase) GetMessagesBetween(_, _, limit int64) []Message {
	messages := []Message{}
	historySize := min(db.size, limit)
	for i := int64(0); i < historySize; i++ {
		messages = append(messages, *NewMessage("", ""))
	}

	return messages
}

func (db *InMemoryDatabase) InsertMessage(m *Message) {
	db.size += 1
}
