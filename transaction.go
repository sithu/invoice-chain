package qbchain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"reflect"
	"time"

	"github.com/izqui/helpers"
)

type Transaction struct {
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
}

type TransactionHeader struct {
	From          []byte
	To            []byte
	Amount        int64
	Timestamp     uint32
	PayloadHash   []byte
	PayloadLength uint32
	Nonce         uint32
}

type TransactionSlice []Transaction

func (slice TransactionSlice) Len() int {

	return len(slice)
}

// Returns bytes to be sent to the network
func NewTransaction(from []byte, to []byte, amount int64, payload []byte) Transaction {

	t := Transaction{
		Header:  TransactionHeader{From: from, To: to, Amount: amount},
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

func (t *Transaction) Sign(keypair *Keypair) []byte {

	s, _ := keypair.Sign(t.Hash())

	return s
}

func (t *Transaction) VerifyTransaction(pow []byte) bool {
	headerHash := t.Hash()
	payloadHash := helpers.SHA256(t.Payload)

	payloadCheck := reflect.DeepEqual(payloadHash, t.Header.PayloadHash)
	powCheck := CheckProofOfWork(pow, headerHash)
	sigCheck := SignatureVerify(t.Header.From, t.Signature, headerHash)
	log.Printf("PayloadCheck:%b, PoWCheck:%b, SigCheck:%b", payloadCheck, powCheck, sigCheck)
	return payloadCheck && powCheck && sigCheck
}

func (t *Transaction) MarshalBinary() ([]byte, error) {

	headerBytes, _ := t.Header.MarshalBinary()

	if len(headerBytes) != TRANSACTION_HEADER_SIZE {
		return nil, errors.New("Transaction Header marshalling error")
	}

	return append(append(headerBytes, helpers.FitBytesInto(t.Signature, NETWORK_KEY_SIZE)...), t.Payload...), nil
}

func (t *Transaction) UnmarshalBinary(d []byte) ([]byte, error) {

	buf := bytes.NewBuffer(d)

	if len(d) < TRANSACTION_HEADER_SIZE+NETWORK_KEY_SIZE {
		return nil, errors.New("Insuficient bytes for unmarshalling transaction")
	}

	header := &TransactionHeader{}
	if err := header.UnmarshalBinary(buf.Next(TRANSACTION_HEADER_SIZE)); err != nil {
		return nil, err
	}

	t.Header = *header

	t.Signature = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	t.Payload = buf.Next(int(t.Header.PayloadLength))

	return buf.Next(helpers.MaxInt), nil

}

func (th *TransactionHeader) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	buf.Write(helpers.FitBytesInto(th.From, NETWORK_KEY_SIZE))
	buf.Write(helpers.FitBytesInto(th.To, NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(helpers.FitBytesInto(th.PayloadHash, 32))
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil

}

func (th *TransactionHeader) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)
	th.From = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	th.To = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Timestamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
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

func (slice TransactionSlice) AddTransaction(t Transaction) TransactionSlice {

	// Inserted sorted by timestamp
	for i, tr := range slice {
		if tr.Header.Timestamp >= t.Header.Timestamp {
			return append(append(slice[:i], t), slice[i:]...)
		}
	}

	return append(slice, t)
}

func (slice *TransactionSlice) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	for _, t := range *slice {

		bs, err := t.MarshalBinary()

		if err != nil {
			return nil, err
		}

		buf.Write(bs)
	}

	return buf.Bytes(), nil
}

func (slice *TransactionSlice) UnmarshalBinary(d []byte) error {

	remaining := d

	for len(remaining) > TRANSACTION_HEADER_SIZE+NETWORK_KEY_SIZE {
		t := new(Transaction)
		rem, err := t.UnmarshalBinary(remaining)

		if err != nil {
			return err
		}
		(*slice) = append((*slice), *t)
		remaining = rem
	}
	return nil
}
