package kvstore

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	db, err := NewBoltDBStore(fmt.Sprintf("/tmp/test_%d.db", time.Now().UnixNano()), "bucketName")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(db.Path())

	keyTest := []byte("key-testing")
	valTest := []byte("value-testing")

	if err := db.Set(keyTest, valTest); err != nil {
		t.Fatal(err)
	}

	val, err := db.Get(keyTest)
	if err != nil {
		t.Fatal(err)
	}

	if string(val) != string(valTest) {
		t.Fatalf("expected key=%q with value=%q, but got: %q", keyTest, valTest, val)
	}
}
