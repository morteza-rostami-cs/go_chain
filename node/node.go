package node

import (
	"context"
	"encoding/json"
	"fmt"
	database "go_chain/database"
	"net/http"
	"time"
)

// check fo mempool txs, every x second => ticker
const miningIntervalTicker = 10 * time.Second

// http port
const DefaultIP = "127.0.0.1"
const DefaultHttpPort = 8081

// node send it's address as parameter to get registered
// GET /node/peer?ip=127.0.0.1&port=8081
const endpointAddPeer = "/node/peer"
const endpointAddPeerQueryKeyIP = "ip"
const endpointAddPeerQueryKeyPort = "port"

const endpointStatus = "/node/status"
const endpointSync = "/node/sync"
const endpointSyncQueryKeyFromBlock = "fromBlock"

// peer nodes, nodes our node knows about
/*
ip and port => to reach the node
*/ 
type PeerNode struct{
	IP	string `json:"ip"`
	Port uint64 `json:"port"`
	// a bootstrap node, is an entry point into the network
	IsBootstrap bool `json:"is_bootstrap"`
	// is this node an active participant, in network
	// IsActive bool `json:"is_active"`

	//private, no json
	connected	bool //our we connected to this peer?
}

// constructor
func NewPeerNode(
	ip string, 
	port uint64, 
	isBootstrap bool,
	connected bool,
) PeerNode {
	return PeerNode{ip, port, isBootstrap, connected}
}
// PeerNode methods

// format the address: localhost:8081
func (self PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", self.IP, self.Port)
}

// our current node object
type Node struct{
	dataDir string// path to /data
	state	*database.State
	// knownPeers	[]PeerNode
	// use map cause: easy to check if a key/peer exists.
	knownPeers map[string]PeerNode

	// transactions waiting to be mined

	// node as a peer information, eg: ip, port
	info	PeerNode

	// archived/mined transactions
	// tx_hash:Tx => key:value
	pendingTXs	map[string]database.Tx // mempool
	archivedTXs	map[string]database.Tx

	// signals coming form network

	// we get the mined block from peer (http), we use this channel, to send a signal to our miner, to stop!
	// we use channel, cause: miner is inside a goroutine.
	newSyncedBlocks	chan	database.Block
	// channel:: receive pendingTxs, that other nodes are registering and broadcasting
	newPendingTXs	chan	database.Tx

	// is node mining or not? => whether the node mining loop is running ?
	isMining	bool
}

// Node constructor
// a fresh node even, needs a peer
// for now: every new node gets a bootstrap 8081
func New(
	dataDir string, 
	ip	string,
	port uint64, 
	bootstrap PeerNode,
) *Node {

	var peers = make(map[string]PeerNode)
	peers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir: dataDir,
		info: NewPeerNode(ip, port, false, false),
		// state: is stored inside Run() method
		knownPeers: peers,

		// these states, only need to be init with empty values
		pendingTXs: make(map[string]database.Tx),
		archivedTXs: make(map[string]database.Tx),
		newSyncedBlocks: make(chan database.Block),

		// this is buffered channel for 10000 txs
		newPendingTXs: make(chan database.Tx, 10_000),
		isMining: false,
	}
}

// Node method

// node peer helpers

func (self *Node) AddPeer(peer PeerNode) {
	// overwrite the key:value, if we have it
	self.knownPeers[peer.TcpAddress()] = peer
}

func (self *Node) RemovePeer(peer PeerNode) {
	//  		 	map				  key
	delete(self.knownPeers, peer.TcpAddress())
}

// check if we know a peer
func (self *Node) IsKnownPeer(peer PeerNode) bool {
	// check for itself
	if peer.IP == self.info.IP && peer.Port == self.info.Port {
		return true
	}

	// check if we have it in knownPeers map
	_, isKnowPeer := self.knownPeers[peer.TcpAddress()]
	return isKnowPeer
}

// decide if to add tx to mempool or not? and maybe do add
func (self *Node) AddPendingTX(
	tx database.Tx,
	fromPeer PeerNode,// use this only for printing name
) error {
	// hash the tx
	txHash, err := tx.Hash()
	if err != nil { return err }

	// we convert to json, to print the tx content, that is it!
	txJson, err := json.Marshal(tx)
	if err != nil { return err }

	// decision:: to add or not to add?
	// if tx is in memPool or archive, do not add again.
	_, isAlreadyPending := self.pendingTXs[txHash.Hex()]
	_, isArchived := self.archivedTXs[txHash.Hex()]

	if !isAlreadyPending && !isArchived {
		fmt.Printf(
			"added pending tx %s from peer %s\n",
			txJson,
			fromPeer.TcpAddress(),
		)

		// add tx to mempool
		self.pendingTXs[txHash.Hex()] = tx
		// send tx through the channel => so other nodes also can add the tx to their mempool
		self.newPendingTXs <- tx
	}

	return nil
}

// memPool transactions map => return as an slice
// for sending over http, we want transactions in a slice
func (self *Node) getPendingTXsAdArray() []database.Tx {
	var txs = make([]database.Tx, 0)// empty slice
	// loop over the map, skip the key/hash
	for _, tx := range self.pendingTXs {
		txs = append(txs, tx)
	}
	return txs
}

/*
# pack pending txs from mempool, into a block
# run the miner on block
# save the mine block, in state.Balances and block.db
*/
func (self *Node) minePendingTXs(ctx context.Context) error {

	// pack mempool txs into a block
	var blockToMine = NewPendingBlock(
		self.state.LatestBlockHash(),// parent
		self.state.NextBlockNumber(), // blockNumber
		// pack all currently available txs inside of the mempool
		self.getPendingTXsAdArray(),
	)

	// mine the pending block
	minedBlock, err := Mine(ctx, blockToMine)
	if err != nil { return err }

	// add mined block to block.db
	_, err = self.state.AddBlock(minedBlock)
	if err != nil { return err }

	// remove from pool and archive
	self.removeMinedPendingTXs(minedBlock)

	return nil
}

// remove tx from mempool and archive it
func (self *Node) removeMinedPendingTXs(block database.Block) {

}

// goroutine:
// runs a loop and check for mempool txs to mine
func (self *Node) mine(ctx context.Context) error {
	// sends a signal through channel ticker.C, every x seconds
	ticker := time.NewTicker(miningIntervalTicker)

	defer ticker.Stop()// later clean up

	// every x seconds, we check for txs and spin up a new goroutine miner
	for {
		select{
		case <-ticker.C:
			// goroutine: every x second, run a goroutine per miner process, that checks for tx mempool
			go func() {
				// if: mempool has txs
				// also: if we already mining: return
				if len(self.pendingTXs) > 0 && !self.isMining {
					// set the node to mining
					self.isMining = true
					// defer the shutdown of miner
					defer func() { self.isMining = false }()

					// make pendingBlock, mine, add block
					err := self.minePendingTXs(ctx)
					if err != nil {
						fmt.Printf("mining error: %s\n", err)
					}
				}
			}()
		// node shutdown clean up
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}

}

// we pass one context into our Run() for all goroutines, to cancel everything on http server shutdown
func (self *Node) Run(ctx context.Context) error {

	//root context -- context is for carrying signals into goroutines. eg: when to cancel the goroutine.
	// ctx := context.Background()
	fmt.Printf("Listening on: %s:%d\n", self.info.IP, self.info.Port)

	// replay the blocks and set the state.Balances and also create /data files, if they do not exist
	// node, replays blocks/transactions, only once (on start up)
	state, err := database.NewStateFromDisk(self.dataDir)
	if err != nil { return err }

	defer state.Close()// close block.db file

	// store a pointer to state on node
	self.state = state // set the state before .sync()

	// print blockchain state
	fmt.Println("Blockchain state:")
	fmt.Printf("	- height: %d\n", self.state.LatestBlock().Header.Number)
	fmt.Printf("	- hash: %s\n", self.state.LatestBlockHash().Hex())

	// goroutine: check for txs in mempool every x second, and spin up miner goroutines
	go self.mine(ctx)

	// sync() is a goroutine that runs in background for syncing network nodes together, we also: pass a context in there.
	go self.sync(ctx) // for each node, we run one goroutine

	// http routes

	http.HandleFunc(
		"/balances/list",
		// router handler
		func(res http.ResponseWriter, req *http.Request) {
			listBalancesHandler(res, req, state)
		},
	)

	http.HandleFunc(
		"/tx/add",
		func(res http.ResponseWriter, req *http.Request) {
			txAddHandler(res, req, self)
		},
	)

	// GET /node/status
	http.HandleFunc(
		endpointStatus,
		func(res http.ResponseWriter, req *http.Request){
			// pass the entire node
			statusHandler(res, req, self)
		},
	)

	// Get /node/sync
	http.HandleFunc(
		endpointSync,
		func(res http.ResponseWriter, req *http.Request) {
			syncHandler(res, req, self.dataDir)
		},
	)

	// GET /node/sync?ip=
	http.HandleFunc(
		endpointAddPeer,
		func(res http.ResponseWriter, req *http.Request){
			addPeerHandler(res, req, self)
		},
	)

	// start the server
	// return http.ListenAndServe(
	// 	// address listen on port :8080, on all available network interfaces
	// 	fmt.Sprintf("%s:%d", self.info.IP, self.info.Port),
	// 	nil)// nil means "use http.DefaultServeMux" — the global mux we just registered our two handlers into.

	// get a http server pointer
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", self.info.Port),
	}

	// this is our control, over shutting down of the http server.
	go func() {
		// block this goroutine, until global cancellation
		<-ctx.Done()
		// then close server
		_ = server.Close()
	}()

	// this does not return, until .Close(), this blocks and keeps the http server alive.
	return server.ListenAndServe()
}