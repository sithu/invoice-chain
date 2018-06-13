package qbchain

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	namespace = []byte("ns")
	key       = []byte("foo")
	data      = []byte("bar")
)

func makeTestDB(t *testing.T) (*DB, func()) {
	tmpDir, _ := ioutil.TempDir("/Users/jduan1/qbchain/", "qbchain-test")

	fmt.Print(tmpDir)

	db, _ := New(path.Join(tmpDir, "data"), path.Join(tmpDir, "meta"))

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
	return db, cleanup
}

func TestDaoSetAndGet(t *testing.T) {
	require := require.New(t)
	db, cleanup := makeTestDB(t)
	defer cleanup()

	err := db.Set(namespace, key, data)
	require.NoError(err)
	storedData, err := db.Get(namespace, key)
	require.NoError(err)
	require.Equal(data, storedData)
}
