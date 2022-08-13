package database

type InMemoryDatabase struct {
	size int
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{}
}

func (db *InMemoryDatabase) GetMessagesBetween(int64, int64, int64) []Message {
	messages := []Message{}
	for i := 0; i < db.size; i++ {
		messages = append(messages, *NewMessage("", ""))
	}

	return messages
}

func (db *InMemoryDatabase) InsertMessage(m *Message) {
	db.size += 1
}
