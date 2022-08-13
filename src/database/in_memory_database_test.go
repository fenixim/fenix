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

	t.Run("insert message length", func(testing *testing.T) {
		for i := 1; i < 3; i++ {
			db := database.NewInMemoryDatabase()
			for j := 0; j < i; j++ {
				db.InsertMessage(
					database.NewMessage("gopher", "hello"))
			}

			got := len(db.GetMessagesBetween(0, time.Now().Unix(), 50))
			expected := i
			test_utils.AssertEqual(t, got, expected)
		}
	})

	t.Run("message content", func(testing *testing.T) {
		testCases := []string{"hello", "yay"}

		for _, test := range(testCases) {
			db := database.NewInMemoryDatabase()
			db.InsertMessage(database.NewMessage("gopher", test))

			history := db.GetMessagesBetween(0, time.Now().Unix(), 1)
			got := history[0].Content
			expected := test
			test_utils.AssertEqual(t, got, expected)
		}
	})
}
