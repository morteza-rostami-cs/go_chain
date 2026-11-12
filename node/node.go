package node

import (
	"bytes"
	"context"
	"encoding/json"
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
// PeerNode methods

// format the address: localhost:8081
func (self PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", self.IP, self.Port)
}

// our current node object
type Node struct{
	dataDir string// path to /data
	port	uint64
	state	*database.State
	// knownPeers	[]PeerNode
	// use map cause: easy to check if a key/peer exists.
	knownPeers map[string]PeerNode
}

// Node constructor
// a fresh node even, needs a peer
// for now: every new node gets a bootstrap 8081
func New(dataDir string, port uint64, bootstrap PeerNode) *Node {

	var peers = make(map[string]PeerNode)
	peers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir: dataDir,
		port: port,
		// state: is stored inside Run() method
		// slice of PeerNodes, init with one
		// knownPeers: []PeerNode{bootstrap},
		knownPeers: peers,
	}
}

// Node method
func (self *Node) Run() error {

	//root context -- context is for carrying signals into goroutines. eg: when to cancel the goroutine.
	ctx := context.Background()
	fmt.Printf("Listening on http port: %d\n", self.port)

	// replay the blocks and set the state.Balances and also create /data files, if they do not exist
	// node, replays blocks/transactions, only once (on start up)
	state, err := database.NewStateFromDisk(self.dataDir)
	if err != nil { return err }

	defer state.Close()// close block.db file

	// store a pointer to state on node
	self.state = state // set the state before .sync()

	// register with bootstrap node
	// if it is not 8081
	if self.port != DefaultHttpPort {
		err := self.registerWithBootstrap()
		if err != nil { return err }
	}

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

	// POST /node/register
	http.HandleFunc(
		"/node/register",
		func(res http.ResponseWriter, req *http.Request) {
			registerPeerHandler(res, req, self)
		},
	)
	
	// start the server
	return http.ListenAndServe(
		fmt.Sprintf(":%d", self.port),// address listen on port :8080, on all available network interfaces
		nil)// nil means "use http.DefaultServeMux" — the global mux we just registered our two handlers into.
}

// nodes can all this method to register with bootstrap, knownPeers
func (self *Node) registerWithBootstrap() error {
	// get bootstrap node 
	// all nodes have: 8081
	bootstrap := self.knownPeers[fmt.Sprintf("127.0.0.1:%d", DefaultHttpPort)]

	// create sender node peer into
	myPeer := NewPeerNode(
		"127.0.0.1",
		self.port,
		false,// sender is not a bootstrap node
		true,
	)

	// convert PeerNode to json, to send over http
	body, err := json.Marshal(myPeer)
	if err != nil { return err }

	// send registration request
	url := fmt.Sprintf("http://%s/node/register", bootstrap.TcpAddress())

	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil { return err }

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed: %s", res.Status)
	}

	fmt.Println("successfully registered with bootstrap node.")
	return nil
}