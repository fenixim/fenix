package database_test

import (
	"fenix/src/database"
	"fenix/src/test_utils"
	"testing"
	"time"
)

type testCase struct {
	author  string
	content string
}

func TestBasicOperations(t *testing.T) {
	t.Run("empty message history", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()

		got, _ := db.GetMessagesBetween(0, time.Now().UnixNano(), 50)
		if got != nil {
			t.Errorf("got %v want nil", got)
		}
	})

	t.Run("message history within limit", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()
		db.InsertMessage(database.NewMessage("gopher", "hello"))

		history, _ := db.GetMessagesBetween(0, time.Now().UnixNano(), 0)
		got := len(history)
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

			history, _ := db.GetMessagesBetween(0, time.Now().UnixNano(), 50)
			got := len(history)
			expected := i
			test_utils.AssertEqual(t, got, expected)
		}
	})

	t.Run("limit takes most recent messages", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()
		db.InsertMessage(database.NewMessage("gopher", "yay"))
		db.InsertMessage(database.NewMessage("kryptic", "fair"))

		history, _ := db.GetMessagesBetween(0, time.Now().UnixNano(), 1)
		got := history[0].Content
		expected := "fair"
		test_utils.AssertEqual(t, got, expected)
	})
}

func TestMessages(t *testing.T) {
	testCases := []testCase{
		{"gopher", "hello"},
		{"bloblet", "yay"},
	}

	for _, test := range testCases {
		db := database.NewInMemoryDatabase()
		db.InsertMessage(
			database.NewMessage(test.author, test.content))

		history, _ := db.GetMessagesBetween(0, time.Now().UnixNano(), 1)

		t.Run("message content", func(testing *testing.T) {
			got := history[0].Content
			expected := test.content
			test_utils.AssertEqual(t, got, expected)
		})

		t.Run("message author", func(testing *testing.T) {
			got := history[0].Author
			expected := test.author
			test_utils.AssertEqual(t, got, expected)
		})
	}
}

func TestTimestamps(t *testing.T) {
	db := database.NewInMemoryDatabase()
	db.InsertMessage(database.NewMessage("gopher", "hello"))
	stamp1 := time.Now().UnixNano()
	time.Sleep(10 * time.Millisecond)
	db.InsertMessage(database.NewMessage("bloblet", "yay"))

	t.Run("messages before timestamp length", func(t *testing.T) {
		history, _ := db.GetMessagesBetween(0, stamp1, 2)
		got := len(history)
		expected := 1

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("messages after timestamp length", func(t *testing.T) {
		history, _ := db.GetMessagesBetween(stamp1, time.Now().UnixNano(), 2)

		got := len(history)
		expected := 1

		test_utils.AssertEqual(t, got, expected)
	})
}
