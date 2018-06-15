package qbchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"time"

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

	go db.runGC()

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
func (db *DB) Get(namespace, key []byte) (value []byte, err error) {
	err = db.badger.View(func(txn *badgerdb.Txn) error {
		item, err := txn.Get(badgerKey(namespace, key))
		if err != nil {
			return err
		}
		value, err = item.Value()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
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

// runGC triggers the garbage collection for the Badger backend db.
func (db *DB) runGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := db.badger.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}

type ChainInfo struct {
	CompanyID string
	Balance   int64
}

func (db *DB) writeChainInfoToDB(bc *Blockchain, namespace []byte) {
	var chainInfo ChainInfo

	// write block to db if not the first dummy block
	if len((*bc.chain.LastBlock().TransactionSlice)) > 0 {
		t := (*bc.chain.LastBlock().TransactionSlice)[0]
		key := t.Header.From
		value, _ := db.Get(namespace, key)
		data := ChainInfo{t.Header.CompanyID, bc.balance}
		byteValue, _ := json.Marshal(data)
		if value == nil {
			log.Printf("create new chain info")
			db.Set(namespace, key, byteValue)
		} else {
			json.Unmarshal(value, &chainInfo)
			chainInfo.Balance = bc.balance
			newValue, _ := json.Marshal(chainInfo)
			log.Printf("update chain info")
			db.Set(namespace, key, newValue)
		}
	}
}

func (db *DB) getChainInfo(pk string, namespace []byte) (chainInfo ChainInfo, err error) {
	value, err := db.Get(namespace, []byte(pk))
	json.Unmarshal(value, &chainInfo)

	return chainInfo, err
}

func MakeDB() (*DB, func()) {
	// dbDir, _ := ioutil.TempDir(".", "qbchain.db")

	dbDir := "./qbchain.db"
	log.Printf(dbDir)

	db, _ := New(path.Join(dbDir, "data"), path.Join(dbDir, "meta"))

	cleanup := func() {
		db.Close()
		//os.RemoveAll(tmpDir)
	}
	return db, cleanup
}

func (db *DB) addBlock(bc *Blockchain, namespace []byte) error {
	Block := *bc.chain.LastBlock()
	t := (*bc.chain.LastBlock().TransactionSlice)[0]
	pk := t.Header.From
	txnID := t.Header.TransactionID
	key := fmt.Sprintf(string(pk) + "_" + txnID)
	blockByte, err := json.Marshal(Block)
	db.Set(namespace, []byte(key), blockByte)
	return err
}
