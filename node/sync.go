package node

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// we call this to get status and peers from other nodes, specially bootstrap_node
// const endpointStatus = "/node/status"

// go routine loop for, periodic sync
func (self *Node) sync(ctx context.Context) error {

	// ticker is a channel, that sends time after each tick
	// 45 * time.Second => 45 sec in nano seconds
	// sends a signal, through a channel, every 45 sec
	/*
	Ticker => sends time, through ticker.C channel.
	you could think of it as another goroutine that send time every 45 sec here inside sync() goroutine.
	*/ 
	ticker := time.NewTicker(45 * time.Second)
	
	// infinite loop,
	for {
		// select: wait for result of multiple channels
		// wait for channel value :: then execute case.
		select {
		//channel blocks and sync() waits for ticker.c value, which comes every 45 sec, then we fetch new blocks and peers
		// <-ticker.C :: value/time is discarded!!
		case <-ticker.C:
			fmt.Println("searching for new peers and blocks...")
			// try fetching new blocks and peers
			self.fetchNewBlocksAndPeers()
		//ctx is a channel -- <-ctx.Done() -- sends cancellation signal, by closing the channel
		case <-ctx.Done():
			// stop the ticker
			ticker.Stop()
			// return error, on cancel
			return ctx.Err()// break, loop and sync()
		}
	}
	// return nil
}

func (self *Node) fetchNewBlocksAndPeers() {

	// get the bootstrap peer, 8081
	// for now: everyone has 8081
	for _, knownPeer := range self.knownPeers {

		// status{Hash, Number, KnownPeers}
		status, err := queryPeerStatus(knownPeer)
		if err != nil {
			fmt.Printf("ERROR: %\n", err)
			continue // check the next peer
		}

		// check if chain has a block
		var newBlocksCount uint64 = 0
		var localBlockCount uint64 = 0
		if self.state.LatestBlock() != nil {

			// local node block number
			// add 1 to get block count, if chain not empty (our block numbers are zero-based)
			localBlockCount = self.state.LatestBlock().Header.Number + 1
	
		} else {
			fmt.Println("chain is empty.")
		}

		// calc block difference
		if localBlockCount < status.Number {
			// difference
			newBlocksCount = status.Number - localBlockCount
			fmt.Printf("Found %d new blocks form peer %s\n", newBlocksCount, knownPeer.TcpAddress())
		}



		// search for new peers in other nodes and add them
		/*
		if bootstrap_node has peers=[]
		others have :: peers=[8081]
		*/ 
		for _, maybeNewPeer := range status.KnownPeers {
			// do i have it?
			_, isKnownPeer := self.knownPeers[maybeNewPeer.TcpAddress()]

			if !isKnownPeer {
				// found a new node
				fmt.Printf("found new peer %s\n", knownPeer.TcpAddress())
				// add new peer
				self.knownPeers[maybeNewPeer.TcpAddress()] = maybeNewPeer
			}
		}
	}
}

// fetch: /node/status data
func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	// create a url
	// http://127.0.0.1:8081/node/status
	url := fmt.Sprintf("http://%s/%s", peer.TcpAddress(), endpointStatus)

	// http request to the other node
	res, err := http.Get(url)
	if err != nil {
		return StatusRes{}, err
	}

	// read json response.body int go type
	statusRes := StatusRes{}
	err = readRes(res, &statusRes)
	if err != nil {
		return StatusRes{}, err
	}

	return statusRes, nil
}