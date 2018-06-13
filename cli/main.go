package main

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/izqui/helpers"
	"github.com/tv42/base58"
)

const (
	KeySize        = 28
	NetworkKeySize = 80
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
		CreateNewTransactionFromCli()
		os.Exit(0)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

// FIXME: duplicate of crypto.go
func generateKeypair() {
	pk, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	b := bigJoin(KeySize, pk.PublicKey.X, pk.PublicKey.Y)

	public := base58.EncodeBig([]byte{}, b)
	private := base58.EncodeBig([]byte{}, pk.D)

	fmt.Printf("Public Key : %s\nPrivate Key: %s\n", public, private)
}

func (k *Keypair) Sign(hash []byte) ([]byte, error) {
	d, err := base58.DecodeToBig(k.Private)
	if err != nil {
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

func CreateNewTransactionFromCli() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Public Key: ")
	publicKey, _ := reader.ReadString('\n')

	fmt.Print("Enter Private Key: ")
	privateKey, _ := reader.ReadString('\n')

	fmt.Print("From Address: ")
	from, _ := reader.ReadString('\n')

	fmt.Print("To Address: ")
	to, _ := reader.ReadString('\n')

	fmt.Print("Payload : ")
	payload, _ := reader.ReadString('\n')

	txn := NewTransaction([]byte(from), []byte(to), []byte(payload))
	kp := Keypair{Public: []byte(publicKey), Private: []byte(privateKey)}
	signature := txn.Sign(&kp)
	txn.Signature = signature

	fmt.Println(txn)
}

type Transaction struct {
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
}

type TransactionHeader struct {
	From          []byte
	To            []byte
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}

// Returns bytes to be sent to the network
func NewTransaction(from, to, payload []byte) *Transaction {
	t := Transaction{Header: TransactionHeader{From: from, To: to}, Payload: payload}
	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = helpers.SHA256(t.Payload)
	t.Header.PayloadLength = uint32(len(t.Payload))
	return &t
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
