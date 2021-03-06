package qbchain

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/izqui/helpers"
	"github.com/spf13/viper"
)

func NewHandler(nodeID string, db *DB) http.Handler {
	h := handler{nil, nodeID, db}

	mux := http.NewServeMux()
	mux.HandleFunc("/nodes/register", buildResponse(h.RegisterNode))
	mux.HandleFunc("/nodes/resolve", buildResponse(h.ResolveConflicts))
	mux.HandleFunc("/transactions/new", buildResponse(h.AddTransaction))
	mux.HandleFunc("/mine", buildResponse(h.Mine))
	mux.HandleFunc("/chain", buildResponse(h.Blockchain))
	return mux
}

type handler struct {
	blockchain *Blockchain
	nodeID     string
	db         *DB
}

type response struct {
	value      interface{}
	statusCode int
	err        error
}

func buildResponse(h func(io.Writer, *http.Request) response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := h(w, r)
		msg := resp.value
		if resp.err != nil {
			msg = resp.err.Error()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.statusCode)
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			log.Printf("could not encode response to output: %v", err)
		}
	}
}

func (h *handler) AddTransaction(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodPost {
		return response{
			nil,
			http.StatusMethodNotAllowed,
			fmt.Errorf("method %s not allowd", r.Method),
		}
	}

	log.Printf("Adding transaction to the blockchain...\n")

	var t Transaction
	err := json.NewDecoder(r.Body).Decode(&t)
	status := http.StatusCreated
	var resp map[string]interface{}

	if err != nil {
		status = http.StatusInternalServerError
		log.Printf("there was an error when trying to add a transaction %v\n", err)
		err = fmt.Errorf("fail to add transaction to the blockchain")
	} else {
		t.Header.Timestamp = uint32(time.Now().Unix())
		t.Header.PayloadHash = helpers.SHA256(t.Payload)
		t.Header.PayloadLength = uint32(len(t.Payload))

		// get blockchain based on pk
		h.blockchain = NewBlockchain(string(t.Header.From), h.db)

		if t.VerifyTransaction(TRANSACTION_POW) {
			block := NewBlock(h.blockchain.latest)
			block.AddTransaction(&t)
			// Hack here, in fact miner should sign the block and add it to chain
			block.BlockHeader.Nonce = t.Header.Nonce
			block.Signature = t.Signature
			block.BlockHeader.Timestamp = t.Header.Timestamp
			block.BlockHash = block.Hash()

			// Forge the new Block by adding it to the chain
			h.blockchain.AddBlock(block, h.db)

			// receiver txn
			rTxn := t
			rTxn.Header.To = t.Header.From
			rTxn.Header.From = t.Header.To
			rTxn.Header.Amount = -t.Header.Amount
			// Write the transacton to the receiver's chain without verification
			rBlockchain := NewBlockchain(string(rTxn.Header.From), h.db)
			rblock := NewBlock(rBlockchain.latest)
			rblock.AddTransaction(&rTxn)
			rblock.BlockHeader.Nonce = rTxn.Header.Nonce
			rblock.Signature = rTxn.Signature
			rblock.BlockHeader.Timestamp = rTxn.Header.Timestamp
			rblock.BlockHeader.Origin = block.BlockHash
			rblock.BlockHash = rblock.Hash()

			// Forge the new Block by adding it to the receiver's chain
			rBlockchain.AddBlock(rblock, h.db)

			// forward the new block to other nodes
			sendToPeers(rblock)
			resp = map[string]interface{}{"message": "New Block Forged", "block": block, "reveiverBlock": rblock}
		} else {
			status = http.StatusBadRequest
			log.Printf("Invalid transaction")
			err = fmt.Errorf("Invalid transaction")
		}

	}

	return response{resp, status, err}
}

func sendToPeers(b Block) {
	peers := viper.GetStringSlice("peer_udp_ports")
	for _, peer := range peers {
		fmt.Println(peer)
		// forward the new block to other nodes
		payload, err := b.MarshalBinary()
		if err != nil {
			fmt.Println("Failed to marshal the block to binary")
		} else {
			SendUDP(payload, peer)
		}
	}
}

func (h *handler) Mine(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodGet {
		return response{
			nil,
			http.StatusMethodNotAllowed,
			fmt.Errorf("method %s not allowd", r.Method),
		}
	}

	log.Println("Mining some coins")

	// We run the proof of work algorithm to get the next proof...
	// lastBlock := h.blockchain.LastBlock()
	// lastProof := lastBlock.Proof
	// proof := h.blockchain.ProofOfWork(lastProof)

	// We must receive a reward for finding the proof.
	// The sender is "0" to signify that this node has mined a new coin.
	newTx := NewTransaction(make([]byte, 0), []byte(h.nodeID), 1, []byte("Mine"))
	prevBlock := h.blockchain.LastBlock()
	block := NewBlock(prevBlock.Hash())
	block.AddTransaction(&newTx)

	block.BlockHeader.Nonce = newTx.Header.Nonce
	block.Signature = newTx.Signature
	block.BlockHeader.Timestamp = newTx.Header.Timestamp

	// Forge the new Block by adding it to the chain
	h.blockchain.AddBlock(block, h.db)

	resp := map[string]interface{}{"message": "New Block Forged", "block": block}
	return response{resp, http.StatusOK, nil}
}

func (h *handler) Blockchain(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodGet {
		return response{
			nil,
			http.StatusMethodNotAllowed,
			fmt.Errorf("method %s not allowd", r.Method),
		}
	}
	log.Println("Blockchain requested")

	pk := r.URL.Query().Get("pk")

	h.blockchain = NewBlockchain(pk, h.db)

	resp := map[string]interface{}{"chain": h.blockchain.chain, "length": len(h.blockchain.chain), "balance": h.blockchain.balance}
	return response{resp, http.StatusOK, nil}
}

func (h *handler) RegisterNode(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodPost {
		return response{
			nil,
			http.StatusMethodNotAllowed,
			fmt.Errorf("method %s not allowd", r.Method),
		}
	}

	log.Println("Adding node to the blockchain")

	var body map[string][]string
	err := json.NewDecoder(r.Body).Decode(&body)

	for _, node := range body["nodes"] {
		h.blockchain.RegisterNode(node)
	}

	resp := map[string]interface{}{
		"message": "New nodes have been added",
		"nodes":   h.blockchain.nodes.Keys(),
	}

	status := http.StatusCreated
	if err != nil {
		status = http.StatusInternalServerError
		err = fmt.Errorf("fail to register nodes")
		log.Printf("there was an error when trying to register a new node %v\n", err)
	}

	return response{resp, status, err}
}

func (h *handler) ResolveConflicts(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodGet {
		return response{
			nil,
			http.StatusMethodNotAllowed,
			fmt.Errorf("method %s not allowd", r.Method),
		}
	}

	log.Println("Resolving blockchain differences by consensus")

	msg := "Our chain is authoritative"
	if h.blockchain.ResolveConflicts() {
		msg = "Our chain was replaced"
	}

	resp := map[string]interface{}{"message": msg, "chain": h.blockchain.chain}
	return response{resp, http.StatusOK, nil}
}
