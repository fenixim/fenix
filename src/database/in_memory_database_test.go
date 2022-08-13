package database_test

import (
	"fenix/src/database"
	"testing"
)

func TestInMemoryDatabase(t *testing.T) {
	t.Run("empty message history", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()

		got := len(db.GetMessagesBetween(0, 0, 50))
		expected := 0

		if got != expected {
			t.Errorf("got %v got %v", got, expected)
		}

	})
}
