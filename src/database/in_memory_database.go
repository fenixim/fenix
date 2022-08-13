package database

type InMemoryDatabase struct {

}

func (*InMemoryDatabase) GetMessagesBetween(int64, int64, int64) []Message {
	return []Message{}
}
