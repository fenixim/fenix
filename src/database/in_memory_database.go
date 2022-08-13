package database

type InMemoryDatabase struct {
	size int64
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{}
}

func (db *InMemoryDatabase) GetMessagesBetween(int64, int64, int64) []Message {
	if db.size == 0 {
		return []Message{}
	} else {
		return []Message{*NewMessage("", "")}
	}
}

func (db *InMemoryDatabase) InsertMessage(m *Message) {
	db.size = 1
}
