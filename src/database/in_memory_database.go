package database

type InMemoryDatabase struct {

}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{}
}

func (*InMemoryDatabase) GetMessagesBetween(int64, int64, int64) []Message {
	return []Message{}
}
