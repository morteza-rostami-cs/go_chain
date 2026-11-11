package node

import (
	"fmt"
	database "go_chain/database"
	"net/http"
)

// http port
const DefaultHttpPort = 8081

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
	IsActive bool `json:"is_active"`
}
// constructor
func NewPeerNode(ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, isActive}
}

// our current node object
type Node struct{
	dataDir string// path to /data
	port	uint64
	state	*database.State
	knownPeers	[]PeerNode
}

// Node constructor
// a fresh node even, needs a peer
func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	return &Node{
		dataDir: dataDir,
		port: port,
		// state: is stored inside Run() method
		// slice of PeerNodes, init with one
		knownPeers: []PeerNode{bootstrap},
	}
}

// Node method
func (self *Node) Run() error {
	// replay the blocks and set the state.Balances and also create /data files, if they do not exist
	// node, replays blocks/transactions, only once (on start up)
	state, err := database.NewStateFromDisk(self.dataDir)
	if err != nil { return err }

	defer state.Close()// close block.db file

	// store a pointer to state on node
	self.state = state

	fmt.Printf("Listening on http port: %d\n", self.port)

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
			txAddHandler(res, req, state)
		},
	)

	// GET /node/status
	http.HandleFunc(
		"/node/status",
		func(res http.ResponseWriter, req *http.Request){
			// pass the entire node
			statusHandler(res, req, self)
		},
	)

	// start the server
	return http.ListenAndServe(
		fmt.Sprintf(":%d", self.port),// address listen on port :8080, on all available network interfaces
		nil)// nil means "use http.DefaultServeMux" — the global mux we just registered our two handlers into.
}