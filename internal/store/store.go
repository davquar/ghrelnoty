package store

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

const ReleasesBucket string = "releases"

type Store struct {
	DB *bolt.DB
}

func Open(path string) (Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return Store{}, fmt.Errorf("cannot open db: %w", err)
	}
	return Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) Get(key string) (string, error) {
	var value []byte
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ReleasesBucket))
		value = b.Get([]byte(key))
		return nil
	})

	return string(value), err
}

func (s *Store) CompareAndSet(key string, value string) (bool, error) {
	var changed bool
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(ReleasesBucket))
		current := b.Get([]byte(key))

		if string(current) == value {
			return nil
		}

		err := b.Put([]byte(key), []byte(value))
		if err != nil {
			return fmt.Errorf("put: %w", err)
		}
		changed = true

		return nil
	})

	return changed, err
}

func (s *Store) Set(key string, value string) error {
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ReleasesBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		err = b.Put([]byte(key), []byte(value))
		if err != nil {
			return fmt.Errorf("put: %w", err)
		}
		return nil
	})
	return err
}
