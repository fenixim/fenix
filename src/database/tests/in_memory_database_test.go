package database_test

import (
	"crypto/rand"
	"fenix/src/database"
	"fenix/src/test_utils"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type testCase struct {
	author  string
	content string
}

func TestBasicOperations(t *testing.T) {
	t.Run("message has id", func(testing *testing.T) {
		db := database.NewInMemoryDatabase()
		msg := database.NewMessage("gopher", "hello")
		db.InsertMessage(msg)

		test_utils.AssertNotEqual(t, msg.MessageID, primitive.NilObjectID)
	})

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
	msg1 := database.NewMessage("gopher", "hello")
	time.Sleep(2 * time.Millisecond)
	msg2 := database.NewMessage("billy", "bye")
	time.Sleep(2 * time.Millisecond)
	msg3 := database.NewMessage("luk", "go")
	time.Sleep(2 * time.Millisecond)
	msg4 := database.NewMessage("josiah", "tdd")

	db.InsertMessage(msg1)
	db.InsertMessage(msg2)
	db.InsertMessage(msg3)
	db.InsertMessage(msg4)

	t.Run("messages before timestamp length", func(t *testing.T) {
		history, _ := db.GetMessagesBetween(0, msg1.Timestamp, 50)
		got := len(history)
		expected := 1

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("messages after timestamp length", func(t *testing.T) {
		history, _ := db.GetMessagesBetween(msg2.Timestamp-int64(time.Millisecond), time.Now().UnixNano(), 50)

		got := len(history)
		expected := 3

		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("messages between timestamps", func(t *testing.T) {
		history, _ := db.GetMessagesBetween(msg2.Timestamp-int64(time.Millisecond), msg3.Timestamp+int64(time.Millisecond), 50)
		got := []string{}

		expected := []string{msg3.Content, msg2.Content}

		for _, m := range history {
			got = append(got, m.Content)
		}

		test_utils.AssertEqual(t, got, expected)
	})
}

func TestUsers(t *testing.T) {

	InsertUser := func(db *database.InMemoryDatabase, username string) *database.User {
		u := &database.User{
			Username: username,
			Password: []byte("myawesomepassword"),
			Salt:     make([]byte, 16),
		}
		rand.Read(u.Salt)

		db.InsertUser(u)
		return u
	}

	t.Run("InsertUser provides userID", func(t *testing.T) {
		db := database.NewInMemoryDatabase()

		u := InsertUser(db, "gopher123")
		got := u.UserID
		dontWant := primitive.NilObjectID

		test_utils.AssertNotEqual(t, got, dontWant)
	})

	t.Run("GetUser prioritizes userID", func(t *testing.T) {
		db := database.NewInMemoryDatabase()
		expected := InsertUser(db, "gopher123")

		got := &database.User{
			UserID:   expected.UserID,
			Username: "NotMyUsername",
		}

		err := db.GetUser(got)

		test_utils.AssertEqual(t, err, nil)
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("GetUser uses username", func(t *testing.T) {
		db := database.NewInMemoryDatabase()
		expected := InsertUser(db, "gopher123")

		got := &database.User{
			Username: expected.Username,
		}

		err := db.GetUser(got)

		test_utils.AssertEqual(t, err, nil)
		test_utils.AssertEqual(t, got, expected)
	})

	t.Run("GetUser with ID returns error", func(t *testing.T) {
		db := database.NewInMemoryDatabase()

		user := &database.User{
			UserID: primitive.NewObjectIDFromTimestamp(time.Now()),
		}

		err := db.GetUser(user)

		test_utils.AssertEqual(t, err, database.DoesNotExist{})
	})

	t.Run("GetUser with Username returns error", func(t *testing.T) {
		db := database.NewInMemoryDatabase()

		user := &database.User{
			Username: "NotMyUsername",
		}

		err := db.GetUser(user)

		test_utils.AssertEqual(t, err, database.DoesNotExist{})
	})
}
