package node

import (
	"encoding/json"
	"go_chain/database"
	"net/http"
)

// http error response
type ErrRes struct{
	Error string `json:"error"`
}

// Get /balances/list response
type BalancesRes struct{
	// the hash of last block
	Hash	database.Hash `json:"block_hash"`
	// map of all account:balance key_values
	Balances map[database.Account]uint `json:"balances"`
}

// POST /create/tx request
type TxAddReq struct{
	From string `json:"from"`
	To string `json:"to"`
	Value uint `json:"value"`
	Data string `json:"data"`
}

// Post /create/tx response
type TxAddRes struct{
	// hash of the last block
	Hash database.Hash `json:"block_hash"`
}

// Get /node/status
// returns: latestBlock.hash and number
// hash:: tells nodes if they have both height=8, do they also agree on chain state (latestBlock.hash)
type StatusRes struct{
	// two nodes, can check this as latest chain state, see if their state matches
	Hash	database.Hash `json:"block_hash"`
	// nodes can check the height, to see if they are behind.
	Number	uint64 `json:"block_number"`
	// p2p nodes (peers)
	// KnownPeers []PeerNode `json:"peers_known"`

	// address: PeerNode map
	// key: 127.0.0.1:8080 and value: PeerNode 
	KnownPeers map[string]PeerNode `json:"peers_known"`
}

type  RegisterPeerRes struct{
	Message	string
}

// http handlers

func listBalancesHandler(res http.ResponseWriter, req *http.Request, state *database.State) {
	// send a json response
	writeRes(
		res, 
		BalancesRes{state.LatestBlockHash(), state.Balances},
	)
}

func txAddHandler(res http.ResponseWriter, req *http.Request, state *database.State){
	// empty request
	request := TxAddReq{}
	// read req from json data
	err := readReq(req, &request)

	if err != nil {
		// format response error
		writeErrRes(res, err)
		return
	}

	// create tx
	tx := database.NewTx(
		database.NewAccount(request.From),
		database.NewAccount(request.To),
		request.Value,
		request.Data,
	)

	// updates :: state.Balances and adds tx to mempool
	err = state.AddTx(tx)
	if err != nil {
		writeErrRes(res, err)
		return
	}

	// persist:: for now we immediately batch txs from mempool into a block and store in block.db (later we change this!!)
	hash, err := state.Persist()
	if err != nil {
		writeErrRes(res, err)
		return
	}

	// final response, convert to json and send http response
	writeRes(res, TxAddRes{hash})
}

func statusHandler(res http.ResponseWriter, req *http.Request, node *Node){
	var blockNumber uint64 = 0// if there is not block yet

	// check for the latest block, because: if block.db is empty, then: sate.LatestBlock() returns: nil
	if lb := node.state.LatestBlock(); lb != nil {
		// we add 1 to (zero-based block numbers to get the block_count)
		blockNumber = lb.Header.Number + 1
	} 

	response := StatusRes{
		Hash: node.state.LatestBlockHash(),
		// this is block count, not block number!
		Number: blockNumber,
		KnownPeers: node.knownPeers,
	}

	writeRes(res, response)
}

// /node/register
func registerPeerHandler(
	res http.ResponseWriter,
	req *http.Request,
	self *Node,	
) {
	var peer PeerNode // Node B as a peer

	// decode json to struct
	err := json.NewDecoder(req.Body).Decode(&peer)
	if err != nil {
		writeErrRes(res , err)
		return
	}

	// add peer into current node.knownPeer
	self.knownPeers[peer.TcpAddress()] = peer

	writeRes(res, RegisterPeerRes{Message: "peer registered"})
}

//=====end handlers===========

