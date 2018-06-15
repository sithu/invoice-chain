package qbchain

import (
	"bytes"
	"encoding/binary"
	// "errors"
	// "reflect"
	// "time"

	"github.com/izqui/helpers"
)

type Block struct {
	*BlockHeader
	Signature []byte
	*TransactionSlice
}

type BlockHeader struct {
	Origin    []byte
	PrevBlock []byte
	Timestamp uint32
	Nonce     uint32
}

type BlockSlice []Block

func (bs BlockSlice) LastBlock() *Block {
	l := len(bs)
	if l == 0 {
		return nil
	} else {
		return &bs[l-1]
	}
}

func (bs *BlockSlice) AppendBlock(b Block) {
	*bs = append(*bs, b)

}

func NewBlock(previousBlock []byte) Block {

	header := &BlockHeader{PrevBlock: previousBlock}
	return Block{header, nil, new(TransactionSlice)}
}

func (b *Block) AddTransaction(t *Transaction) {
	newSlice := b.TransactionSlice.AddTransaction(*t)
	b.TransactionSlice = &newSlice
}

func (b *Block) Sign(keypair *Keypair) []byte {

	s, _ := keypair.Sign(b.Hash())
	return s
}

func (b *Block) VerifyBlock(prefix []byte) bool {

	headerHash := b.Hash()

	return CheckProofOfWork(prefix, headerHash) && SignatureVerify(b.BlockHeader.Origin, b.Signature, headerHash)
}

func (b *Block) Hash() []byte {

	headerHash, _ := b.BlockHeader.MarshalBinary()
	return helpers.SHA256(headerHash)
}

func (b *Block) GenerateNonce(prefix []byte) uint32 {

	newB := b
	for {

		if CheckProofOfWork(prefix, newB.Hash()) {
			break
		}

		newB.BlockHeader.Nonce++
	}

	return newB.BlockHeader.Nonce
}

func (b *Block) MarshalBinary() ([]byte, error) {

	bhb, err := b.BlockHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	sig := helpers.FitBytesInto(b.Signature, NETWORK_KEY_SIZE)
	tsb, err := b.TransactionSlice.MarshalBinary()

	if err != nil {
		return nil, err
	}

	return append(append(bhb, sig...), tsb...), nil
}

func (b *Block) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)

	header := new(BlockHeader)
	err := header.UnmarshalBinary(buf.Next(BLOCK_HEADER_SIZE))
	if err != nil {
		return err
	}

	b.BlockHeader = header
	b.Signature = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)

	ts := new(TransactionSlice)
	err = ts.UnmarshalBinary(buf.Next(helpers.MaxInt))
	if err != nil {
		return err
	}

	b.TransactionSlice = ts

	return nil
}

func (h *BlockHeader) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	buf.Write(helpers.FitBytesInto(h.Origin, NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, h.Timestamp)
	buf.Write(helpers.FitBytesInto(h.PrevBlock, 32))
	binary.Write(buf, binary.LittleEndian, h.Nonce)

	return buf.Bytes(), nil
}

func (h *BlockHeader) UnmarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)
	h.Origin = helpers.StripByte(buf.Next(NETWORK_KEY_SIZE), 0)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &h.Timestamp)
	h.PrevBlock = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &h.Nonce)

	return nil
}
