package kvstore

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCreateDBAsInterface(t *testing.T) {
	var db Store
	var err error

	db, err = NewBoltDBStore("/tmp/test.db")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSetAndGet(t *testing.T) {
	db, err := NewBoltDBStore(fmt.Sprintf("/tmp/test_%d.db", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(db.Path())

	keyTest := []byte("key-testing")
	valTest := []byte("value-testing")

	if err := db.Set(DefaultBucket, keyTest, valTest); err != nil {
		t.Fatal(err)
	}

	val, err := db.Get(DefaultBucket, keyTest)
	if err != nil {
		t.Fatal(err)
	}

	if string(val) != string(valTest) {
		t.Fatalf("expected key=%q with value=%q, but got: %q", keyTest, valTest, val)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNotFoundKey(t *testing.T) {
	db, err := NewBoltDBStore(fmt.Sprintf("/tmp/test_%d.db", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(db.Path())

	val, err := db.Get(DefaultBucket, []byte("non existing key"))
	if err != nil {
		t.Fatal(err)
	}

	if val != nil {
		t.Fatalf("unexpected non-nil value: %q", val)
	}
}
