package qbchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	// "time"
	"log"
)

type BlockchainService interface {
	// Add a new node to the list of nodes
	RegisterNode(address string) bool

	// Determine if a given blockchain is valid
	ValidChain(chain Blockchain) bool

	// This is our Consensus Algorithm, it resolves conflicts by replacing
	// our chain with the longest one in the network.
	ResolveConflicts() bool

	// Create a new Block in the Blockchain
	AddBlock(b Block)

	// Creates a new transaction to go into the next mined Block
	NewTransaction(tx Transaction) int64

	// Returns the last block on the chain
	LastBlock() *Block

	// Simple Proof of Work Algorithm:
	// - Find a number p' such that hash(pp') contains leading 4 zeroes, where p is the previous p'
	// - p is the previous proof, and p' is the new proof
	ProofOfWork(lastProof int64)

	// Validates the Proof: Does hash(lastProof, proof) contain 4 leading zeroes?
	VerifyProof(lastProof, proof int64) bool
}

type Blockchain struct {
	chain   BlockSlice
	balance int64
	nodes   StringSet
}

func (bc *Blockchain) AddBlock(b Block, db *DB) {

	bc.chain = append(bc.chain, b)
	// Sum all txns balance
	for _, tx := range *b.TransactionSlice {
		bc.balance += tx.Header.Amount
	}

	// save to DB
	db.writeChainInfoToDB(bc, []byte(DB_NAMESPACE))
	db.addBlock(bc, []byte(DB_NAMESPACE))

	if len(*b.TransactionSlice) > 0 {
		t := (*b.TransactionSlice)[0]
		value, _ := db.Get([]byte(DB_NAMESPACE), t.Header.From)
		log.Printf("current balance:" + string(value))
	}

}

func (bc *Blockchain) NewTransaction(tx Transaction) int64 {
	// bc.transactions = append(bc.transactions, tx)
	return 1
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain.LastBlock()
}

func (bc *Blockchain) ProofOfWork(lastProof int64) int64 {
	var proof int64 = 0
	for !bc.ValidProof(lastProof, proof) {
		proof += 1
	}
	return proof
}

func (bc *Blockchain) ValidProof(lastProof, proof int64) bool {
	guess := fmt.Sprintf("%d%d", lastProof, proof)
	guessHash := ComputeHashSha256([]byte(guess))
	return guessHash[:4] == "0000"
}

func (bc *Blockchain) ValidChain(chain *BlockSlice) bool {
	for _, block := range *chain {
		// Check that the hash of the block is correct
		if !block.VerifyBlock(BLOCK_POW) {
			return false
		}
	}
	return true
}

func (bc *Blockchain) RegisterNode(address string) bool {
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	return bc.nodes.Add(u.Host)
}

func (bc *Blockchain) ResolveConflicts() bool {
	neighbours := bc.nodes
	newChain := new(BlockSlice)

	// We're only looking for chains longer than ours
	maxLength := len(bc.chain)

	// Grab and verify the chains from all the nodes in our network
	for _, node := range neighbours.Keys() {
		otherBlockchain, err := findExternalChain(node)
		if err != nil {
			continue
		}

		// Check if the length is longer and the chain is valid
		if otherBlockchain.Length > maxLength && bc.ValidChain(&otherBlockchain.Chain) {
			maxLength = otherBlockchain.Length
			newChain = &otherBlockchain.Chain
		}
	}
	// Replace our chain if we discovered a new, valid chain longer than ours
	if len(*newChain) > 0 {
		bc.chain = *newChain
		return true
	}

	return false
}

func NewBlockchain(pk string, db *DB) *Blockchain {
	value, _ := db.getChainInfo(pk, []byte(DB_NAMESPACE))

	newBlockchain := &Blockchain{
		chain:   nil,
		balance: value.Balance,
		nodes:   NewStringSet(),
	}
	// Initial, sentinel block if not block in DB
	newBlockchain.AddBlock(NewBlock(make([]byte, 0)), db) // empty previous block

	db.getBlocks(newBlockchain, pk+"_", []byte(DB_NAMESPACE))

	return newBlockchain
}

func computeHashForBlock(block Block) string {
	var buf bytes.Buffer
	// Data for binary.Write must be a fixed-size value or a slice of fixed-size values,
	// or a pointer to such data.
	jsonblock, marshalErr := json.Marshal(block)
	if marshalErr != nil {
		log.Fatalf("Could not marshal block: %s", marshalErr.Error())
	}
	hashingErr := binary.Write(&buf, binary.BigEndian, jsonblock)
	if hashingErr != nil {
		log.Fatalf("Could not hash block: %s", hashingErr.Error())
	}
	return ComputeHashSha256(buf.Bytes())
}

type blockchainInfo struct {
	Length  int        `json:"length"`
	Chain   BlockSlice `json:"chain"`
	Balance int        `json:"balance"`
}

func findExternalChain(address string) (blockchainInfo, error) {
	response, err := http.Get(fmt.Sprintf("http://%s/chain", address))
	if err == nil && response.StatusCode == http.StatusOK {
		var bi blockchainInfo
		if err := json.NewDecoder(response.Body).Decode(&bi); err != nil {
			return blockchainInfo{}, err
		}
		return bi, nil
	}
	return blockchainInfo{}, err
}
