package database_test

import (
	"fenix/src/database"
	"fenix/src/test_utils"
	"testing"
	"time"
)

func TestInMemoryDatabase(t *testing.T) {
	t.Run("empty message history", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()

		got := len(db.GetMessagesBetween(0, time.Now().Unix(), 50))
		expected := 0
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("message history within limit", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()
		db.InsertMessage(database.NewMessage("gopher", "hello"))

		got := len(db.GetMessagesBetween(0, time.Now().Unix(), 0))
		expected := 0

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("insert 1 message length", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()
		db.InsertMessage(database.NewMessage("gopher", "hello"))

		got := len(db.GetMessagesBetween(0, time.Now().Unix(), 50))
		expected := 1
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("insert 2 message length", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()
		db.InsertMessage(database.NewMessage("gopher", "hello"))
		db.InsertMessage(database.NewMessage("gopher", "hello"))

		got := len(db.GetMessagesBetween(0, time.Now().Unix(), 50))
		expected := 2
		test_utils.AssertEqual(t, got, expected)
	})
}
