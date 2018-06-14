package qbchain

import (
	"encoding/json"
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

type Test123 struct {
	Proof        int64  `json:"proof"`
	PreviousHash string `json:"previous_hash"`
}

func makeDBTest(t *testing.T) (*DB, func()) {
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
	db, cleanup := makeDBTest(t)
	defer cleanup()

	err := db.Set(namespace, key, data)
	require.NoError(err)
	storedData, err := db.Get(namespace, key)
	require.NoError(err)
	require.Equal(data, storedData)
}

//test value is a json struct
func TestDaoSetAndGetforStruc(t *testing.T) {
	require := require.New(t)
	db, cleanup := makeDBTest(t)
	defer cleanup()

	x := Test123{1, "Hello"}
	// reqBodyBytes := new(bytes.Buffer)
	// json.NewEncoder(reqBodyBytes).Encode(x)
	xByte, _ := json.Marshal(x)
	err := db.Set(namespace, key, xByte)
	require.NoError(err)
	storedData, err := db.Get(namespace, key)
	require.NoError(err)
	require.Equal(xByte, storedData)
}
