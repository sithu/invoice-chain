package main

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/izqui/helpers"
	"github.com/tv42/base58"
)

const (
	KeySize                    = 28
	NetworkKeySize             = 80
	TRANSACTION_POW_COMPLEXITY = 1
	POW_PREFIX                 = 0
	BLOCK_POW_COMPLEXITY       = 2
)

type Keypair struct {
	Public  []byte `json:"public"`
	Private []byte `json:"private"`
}

func main() {
	genkeysCommand := flag.NewFlagSet("genkeys", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("genkeys|submit is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "genkeys":
		genkeysCommand.Parse(os.Args[2:])
		fmt.Println("Generating a key pair...")
		generateKeypair()
		os.Exit(0)
	case "submit":
		txn := CreateNewTransactionFromCli()
		httpPOST(txn)
		os.Exit(0)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

type User struct {
	Id      string
	Balance uint64
}

func httpPOST(t Transaction) {
	// u := User{Id: "US123", Balance: 8}
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(t)
	resp, err := http.Post("http://127.0.0.1:8000/transactions/new", "application/json; charset=utf-8", buffer)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.Status, string(body))
}

// FIXME: duplicate of crypto.go
func generateKeypair() *Keypair {
	pk, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	b := bigJoin(KeySize, pk.PublicKey.X, pk.PublicKey.Y)

	public := base58.EncodeBig([]byte{}, b)
	private := base58.EncodeBig([]byte{}, pk.D)

	fmt.Printf("Public Key : %s\nPrivate Key: %s\n", public, private)
	kp := Keypair{Public: public, Private: private}
	return &kp
}

func (k *Keypair) Sign(hash []byte) ([]byte, error) {
	fmt.Println("Private Key:", k.Private)
	d, err := base58.DecodeToBig(k.Private)
	if err != nil {
		fmt.Printf("Error:%s", err)
		return nil, err
	}

	b, _ := base58.DecodeToBig(k.Public)
	pub := splitBig(b, 2)
	x, y := pub[0], pub[1]

	key := ecdsa.PrivateKey{ecdsa.PublicKey{elliptic.P224(), x, y}, d}
	r, s, _ := ecdsa.Sign(rand.Reader, &key, hash)

	return base58.EncodeBig([]byte{}, bigJoin(KeySize, r, s)), nil
}

// FIXME: duplicate of crypto.go
func bigJoin(expectedLen int, bigs ...*big.Int) *big.Int {
	bs := []byte{}
	for i, b := range bigs {
		by := b.Bytes()
		dif := expectedLen - len(by)

		if dif > 0 && i != 0 {
			by = append(helpers.ArrayOfBytes(dif, 0), by...)
		}

		bs = append(bs, by...)
	}

	b := new(big.Int).SetBytes(bs)
	return b
}

func CreateNewTransactionFromCli() Transaction {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Public Key: ")
	publicKey, _ := reader.ReadString('\n')
	publicKey = strings.TrimSpace(publicKey)

	fmt.Print("Enter Private Key: ")
	privateKey, _ := reader.ReadString('\n')
	privateKey = strings.TrimSpace(privateKey)

	fmt.Print("To Address: ")
	to, _ := reader.ReadString('\n')
	to = strings.TrimSpace(to)

	fmt.Print("Amount: ")
	amount, _ := reader.ReadString('\n')
	amount = strings.TrimSpace(amount)

	amt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	fmt.Print("Company ID: ")
	cid, _ := reader.ReadString('\n')

	fmt.Print("Transaction ID: ")
	tid, _ := reader.ReadString('\n')

	fmt.Print("Payload : ")
	payload, _ := reader.ReadString('\n')
	payload = strings.TrimSpace(payload)

	kp := Keypair{Public: []byte(publicKey), Private: []byte(privateKey)}
	txn := NewTransaction(kp.Public, []byte(to), amt, cid, tid, []byte(payload))
	sig := txn.Sign(&kp)
	txn.Signature = sig
	return txn
}

type Transaction struct {
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
}

type TransactionHeader struct {
	From          []byte
	To            []byte
	Amount        int64
	CompanyID     string
	TransactionID string
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}

// Returns bytes to be sent to the network
func NewTransaction(from []byte, to []byte, amount int64, cid string, tid string, payload []byte) Transaction {
	t := Transaction{
		Header:  TransactionHeader{From: from, To: to, Amount: amount, CompanyID: cid, TransactionID: tid},
		Payload: payload}

	payloadByte := []byte(payload)
	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = helpers.SHA256(payloadByte)
	t.Header.PayloadLength = uint32(len(payloadByte))
	t.Header.Nonce = t.GenerateNonce(TRANSACTION_POW)
	return t
}

func (t *Transaction) Hash() []byte {
	headerBytes, _ := t.Header.MarshalBinary()
	return helpers.SHA256(headerBytes)
}

func (t *Transaction) Sign(keypair *Keypair) []byte {
	s, _ := keypair.Sign(t.Hash())
	return s
}

func (th *TransactionHeader) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(helpers.FitBytesInto(th.From, NetworkKeySize))
	buf.Write(helpers.FitBytesInto(th.To, NetworkKeySize))
	binary.Write(buf, binary.LittleEndian, th.CompanyID)
	binary.Write(buf, binary.LittleEndian, th.TransactionID)
	binary.Write(buf, binary.LittleEndian, th.Amount)
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(helpers.FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil
}

func splitBig(b *big.Int, parts int) []*big.Int {
	bs := b.Bytes()
	if len(bs)%2 != 0 {
		bs = append([]byte{0}, bs...)
	}

	l := len(bs) / parts
	as := make([]*big.Int, parts)

	for i := range as {
		as[i] = new(big.Int).SetBytes(bs[i*l : (i+1)*l])
	}

	return as
}

func (t *Transaction) GenerateNonce(prefix []byte) uint32 {

	newT := t
	for {

		if CheckProofOfWork(prefix, newT.Hash()) {
			break
		}

		newT.Header.Nonce++
	}

	return newT.Header.Nonce
}

var (
	TRANSACTION_POW = helpers.ArrayOfBytes(TRANSACTION_POW_COMPLEXITY, POW_PREFIX)
	BLOCK_POW       = helpers.ArrayOfBytes(BLOCK_POW_COMPLEXITY, POW_PREFIX)
)

func CheckProofOfWork(prefix []byte, hash []byte) bool {

	if len(prefix) > 0 {
		return reflect.DeepEqual(prefix, hash[:len(prefix)])
	}
	return true
}
