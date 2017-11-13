package kvstore

import (
	"fmt"

	"github.com/boltdb/bolt"
)

const DefaultBucket = "bucket-name"

type Store interface {
	Set(bucket string, key []byte, value []byte) error
	Get(bucket string, key []byte) ([]byte, error)
	Close() error
}

// NewBoltDBStore creates a BoltStore pointer otherwise returns nil and an error.
func NewBoltDBStore(filePath string) (*BoltStore, error) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(DefaultBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &BoltStore{DB: db}, nil
}

type BoltStore struct {
	*bolt.DB
}

// Set saves the value under the passed key.
func (b *BoltStore) Set(bucket string, key []byte, value []byte) error {
	return b.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Put(key, value)
	})
}

// Get retrieves the value for a key.
// Returns a nil value if the key does not exist.
func (b *BoltStore) Get(bucket string, key []byte) ([]byte, error) {
	var value []byte

	if err := b.View(func(tx *bolt.Tx) error {
		value = tx.Bucket([]byte(bucket)).Get(key)
		return nil
	}); err != nil {
		return nil, err
	}

	return value, nil
}

// Close closes a bolt DB.
func (b *BoltStore) Close() error {
	return b.DB.Close()
}
