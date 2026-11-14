package node

import (
	// "encoding/json"
	"fmt"
	"go_chain/database"
	"net/http"
	"strconv"
	"time"
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
	Message	string `json:"message"`
}

// GET /node/peer
type AddPeerRes struct{
	Success	bool	`json:"success"`
	Error	string	`json:"error"`
}

// /node/sync?hash=sdfsdf
type SyncRes struct{
	Blocks []database.Block `json:"blocks"`
}

// http handlers

func listBalancesHandler(res http.ResponseWriter, req *http.Request, state *database.State) {
	// send a json response
	writeRes(
		res, 
		BalancesRes{state.LatestBlockHash(), state.Balances},
	)
}

func txAddHandler(
	res http.ResponseWriter, 
	req *http.Request, 
	state *database.State,
){
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

	// make a new block
	var block = database.NewBlock(
		state.LatestBlockHash(),// parent
		state.NextBlockNumber(), // number
		0, // nonce
		uint64(time.Now().Unix()), // time

		// pass our tx into block
		[]database.Tx{tx}, // txs
	) 

	// applies to state.Balances and also persists in block.db
	hash, err := state.AddBlock(block)
	if err != nil {
		writeErrRes(res, err)
		return
	}

	// final response, convert to json and send http response
	writeRes(res, TxAddRes{hash})
}

func statusHandler(
	res http.ResponseWriter, 
	req *http.Request, 
	node *Node,
){

	// check for the latest block, because: if block.db is empty, then: sate.LatestBlock() returns: nil

	response := StatusRes{
		Hash: node.state.LatestBlockHash(),
		// this is block count, not block number!
		Number: node.state.GetChainLen(),
		KnownPeers: node.knownPeers,
	}

	writeRes(res, response)
}

// /node/register
// func registerPeerHandler(
// 	res http.ResponseWriter,
// 	req *http.Request,
// 	self *Node,	
// ) {
// 	var peer PeerNode // Node B as a peer

// 	// decode json to struct
// 	err := json.NewDecoder(req.Body).Decode(&peer)
// 	if err != nil {
// 		writeErrRes(res , err)
// 		return
// 	}

// 	// add peer into current node.knownPeer
// 	self.knownPeers[peer.TcpAddress()] = peer

// 	writeRes(res, RegisterPeerRes{Message: "peer registered"})
// }

// /node/sync
func syncHandler(
	res http.ResponseWriter, 
	req *http.Request, 
	dataDir string,
) {
	// get the hash, from: url ?fromBlock=sdfdss -- query_parameter
	reqHash := req.URL.Query().Get(endpointSyncQueryKeyFromBlock)

	var hash = database.Hash{}
	// string to bytes to Hash 32
	err := hash.UnmarshalText([]byte(reqHash))
	if err != nil {
		writeErrRes(res, err)
		return
	}

	// get all the blocks after block.hash
	blocks, err := database.GetBlocksAfter(hash, dataDir)
	if err != nil {
		writeErrRes(res, err)
		return
	}

	writeRes(res, SyncRes{Blocks: blocks})
}

// a node sends: ip:port => and register with other
func addPeerHandler(
	res http.ResponseWriter,
	req *http.Request,
	node	*Node,
) {
	// gets query params: /node/peer?ip=localhost&port=8081
	var peerIP = req.URL.Query().Get(endpointAddPeerQueryKeyIP)
	var peerPortRaw = req.URL.Query().Get(endpointAddPeerQueryKeyPort)

	// convert string to uint
	peerPort, err := strconv.ParseUint(peerPortRaw, 10, 32)
	if err != nil {
		writeRes(res, AddPeerRes{false, err.Error()})
		return
	}

	// create a new peerNode
	var peer = NewPeerNode(
		peerIP, 
		peerPort, 
		false, // is_bootstrap
		true, // connected
	)

	// add to this node knownPeers
	node.AddPeer(peer) 

	fmt.Printf("peer %s was added into KnownPeers\n", peer.TcpAddress())

	writeRes(res, AddPeerRes{Success: true, Error: ""})
}

//=====end handlers===========

