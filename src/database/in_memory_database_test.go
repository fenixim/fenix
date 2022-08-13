package database_test

import (
	"fenix/src/database"
	"testing"
)

func newInMemoryDatabase() *database.InMemoryDatabase {
	return &database.InMemoryDatabase{}
}

func TestInMemoryDatabase(t *testing.T) {
	t.Run("empty message history", func(testing *testing.T) {
		db := newInMemoryDatabase()

		got := len(db.GetMessagesBetween(0, 0, 50))
		expected := 0

		if got != expected {
			t.Errorf("got %v got %v", got, expected)
		}

	})
}
