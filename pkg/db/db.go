// Implements simple Bolt-based Key Value pair database with Get and Put operations
// TODO: add delete and prefix query support
package db

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
)

type Database struct {
	db *bolt.DB // bolt db
}

func NewDatabase(dbpath string) (db *Database, close func() error, err error) {
	// init bolt database
	_db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		return nil, nil, err
	}

	db = &Database{db: _db}
	close = _db.Close

	// create bucket for kv store
	err2 := db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err2 != nil {
		close()
		return nil, nil, err2
	}

	return db, close, nil
}

func (d *Database) Put(key []byte, val []byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		err := b.Put(key, val)
		return err
	})
}

func (d *Database) Get(key []byte) ([]byte, error) {
	var res []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		res = b.Get(key)
		return nil
	})

	return res, err
}
