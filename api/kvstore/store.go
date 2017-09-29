package kvstore

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type KVStore interface {
	Set(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Close() error
}

func NewBoltDBStore(filePath, bucketName string) (*BoltStore, error) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucketName))
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

func (b *BoltStore) Set(key []byte, value []byte) error {
	return b.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(b.bucket).Put(key, value)
	})
}

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

func (b *BoltStore) Close() error {
	return b.Close()
}
