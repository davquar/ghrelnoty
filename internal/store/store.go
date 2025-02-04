package store

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// ReleasesBucket is the name of the bucket in which release data
// is stored on Bolt.
const ReleasesBucket string = "releases"

// Store holds the instance to the Bolt database.
type Store struct {
	DB *bolt.DB
}

// Open opens (creating it if needed) a new Bolt database file.
func Open(path string) (Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return Store{}, fmt.Errorf("cannot open db: %w", err)
	}
	return Store{DB: db}, nil
}

// Close closes the connection with the database.
func (s *Store) Close() error {
	return s.DB.Close()
}

// Get returns the value of the given key from the database.
func (s *Store) Get(key string) (string, error) {
	var value []byte
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ReleasesBucket))
		value = b.Get([]byte(key))
		return nil
	})

	return string(value), err
}

// CompareAndSet writes the given value for the given key in the database, if the new
// and current values are different, returning true if the change has been done.
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

// Set writes the given value for the given key in the database.
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
