package kvstore

import (
	"fmt"

	"github.com/boltdb/bolt"
)

const DefaultBucket = "bucket-name"

type Store interface {
	Set(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Close() error
}

// NewBoltDBStore creates a BoltStore pointer otherwise returns nil and an error.
func NewBoltDBStore(filePath, bucketName string) (*BoltStore, error) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return nil, err
	}
	if bucketName == "" {
		bucketName = DefaultBucket
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &BoltStore{
		DB:     db,
		bucket: []byte(bucketName),
	}, nil
}

type BoltStore struct {
	*bolt.DB
	bucket []byte
}

// Set saves the value under the passed key.
func (b *BoltStore) Set(key []byte, value []byte) error {
	return b.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(b.bucket).Put(key, value)
	})
}

// Get retrieves the value for a key.
// Returns a nil value if the key does not exist.
func (b *BoltStore) Get(key []byte) ([]byte, error) {
	var value []byte

	if err := b.View(func(tx *bolt.Tx) error {
		value = tx.Bucket(b.bucket).Get(key)
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
