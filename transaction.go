package qbchain

import (
	"bytes"
	"encoding/binary"
	// "errors"
	// "reflect"
	"time"

	"github.com/izqui/helpers"
)

const (
	NETWORK_KEY_SIZE = 80
)

type Transaction struct {
	Header    TransactionHeader  
	Signature []byte
	Payload   string
}

type TransactionHeader struct {
	From          string 
	To            string 
	Amount    	  int64  
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}

// Returns bytes to be sent to the network
func NewTransaction(from string, to string, amount int64, payload string) Transaction {

	t := Transaction{
		Header: TransactionHeader{From: from, To: to, Amount: amount},
		Payload: payload}

	payloadByte := []byte(payload)
	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = helpers.SHA256(payloadByte)
	t.Header.PayloadLength = uint32(len(payloadByte))

	return t
}

func (t *Transaction) Hash() []byte {

	headerBytes, _ := t.Header.MarshalBinary()
	return helpers.SHA256(headerBytes)
}

// func (t *Transaction) Sign(keypair *Keypair) []byte {

// 	s, _ := keypair.Sign(t.Hash())

// 	return s
// }

// func (t *Transaction) VerifyTransaction(pow []byte) bool {

// 	headerHash := t.Hash()
// 	payloadHash := helpers.SHA256(t.Payload)

// 	return reflect.DeepEqual(payloadHash, t.Header.PayloadHash) && CheckProofOfWork(pow, headerHash) && SignatureVerify(t.Header.From, t.Signature, headerHash)
// }

func (th *TransactionHeader) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	buf.Write(helpers.FitBytesInto([]byte(th.From), NETWORK_KEY_SIZE))
	buf.Write(helpers.FitBytesInto([]byte(th.To), NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(helpers.FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil

}

func (th *TransactionHeader) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)
	th.From = string(helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0))
	th.To = string(helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0))
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Timestamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
}