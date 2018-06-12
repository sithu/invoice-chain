package qbchain

import (
	"errors"

	badgerdb "github.com/dgraph-io/badger"
)

type DB struct {
	badger *badgerdb.DB
}

// New creates a new database implementation using badger
// both data and meta are required. they are dir where data where be stored(can use "/tmp/badger")
func New(data, meta string) (*DB, error) {
	if len(data) == 0 {
		return nil, errors.New("no data directory defined")
	}
	if len(meta) == 0 {
		return nil, errors.New("no meta directory defined")
	}
	opts := badgerdb.DefaultOptions
	opts.Dir, opts.ValueDir = meta, data
	badgerDB, err := badgerdb.Open(opts)
	if err != nil {
		return nil, err
	}

	db := &DB{
		badger: badgerDB,
	}
	return db, nil
}

// Set implements db.Set
func (db *DB) Set(namespace, key, metadata []byte) error {
	err := db.badger.Update(func(txn *badgerdb.Txn) error {
		return txn.Set(badgerKey(namespace, key), metadata)
	})
	if err != nil {
		return err
	}
	return nil
}

// Get implements db.Get
func (db *DB) Get(namespace, key []byte) (metadata []byte, err error) {
	err = db.badger.View(func(txn *badgerdb.Txn) error {
		item, err := txn.Get(badgerKey(namespace, key))
		if err != nil {
			return err
		}
		value, err := item.Value()
		if err != nil {
			return err
		}
		metadata = make([]byte, len(value))
		copy(metadata, value)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func badgerPrefix(namespace []byte) []byte {
	return []byte(string(namespace) + "/")
}

func badgerKey(namespace, key []byte) []byte {
	return append(badgerPrefix(namespace), key...)
}

// Close DB
func (db *DB) Close() error {
	// close db
	err := db.badger.Close()
	if err != nil {
		return err
	}
	return nil
}
