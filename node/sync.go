package node

import (
	"context"
	"fmt"
	"go_chain/database"
	"net/http"
	"time"
)

// we call this to get status and peers from other nodes, specially bootstrap_node
// const endpointStatus = "/node/status"

// go routine:: loop for, periodic sync
func (self *Node) sync(ctx context.Context) error {

	// ticker is a channel, that sends time after each tick
	// 45 * time.Second => 45 sec in nano seconds
	// sends a signal, through a channel, every 45 sec
	/*
	Ticker => sends time, through ticker.C channel.
	you could think of it as another goroutine that send time every 45 sec here inside sync() goroutine.
	*/ 
	ticker := time.NewTicker(8 * time.Second)
	defer ticker.Stop()// if sync done! clean up ticker
	
	// infinite loop,
	for {
		// select: wait for result of multiple channels
		// wait for channel value :: then execute case.
		select {
		//channel blocks and sync() waits for ticker.c value, which comes every 45 sec, then we fetch new blocks and peers
		// <-ticker.C :: value/time is discarded!!
		case <-ticker.C:
			self.doSync()
			// fmt.Println("searching for new peers and blocks...")
			// try fetching new blocks and peers
			// self.fetchNewBlocksAndPeers()
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

// push my to bootstrap
// pull peers from bootstrap
// pull missing blocks from each peers
func (self *Node) doSync() {

	// initially we only have 8081
	for _, peer := range self.knownPeers {

		// skip empty ip
		if peer.IP == "" {
			continue
		}

		// skip self
		// do not register with yourself
		// do not pull peers from yourself
		// do not pull blocks from yourself
		if self.info.IP == peer.IP && self.info.Port == peer.Port {
			fmt.Println("===== skip my self=====")
			continue
		}

		fmt.Printf("\nsearching from node: %s for peers and blocks.\n", peer.TcpAddress())

		// status{hash, number, knownPeers}
		status, err := queryPeerStatus(peer)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)

			// if it fails to query /node/status, remove it
			// we assume peer is not reachable (dead)
			fmt.Printf("\npeer %s was removed from knownPeers\n", peer.TcpAddress())
			// peer unreachable!! if it comes online, it gets added later, for now just fuck off!
			self.RemovePeer(peer)

			continue // dead peer, move on
		}

		// we join 8081 (register with bootstrap, and others)		
		err = self.joinKnownPeers(peer)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			// we can not join, continue next peer
			// we assume that /node/peer is not accessible, and we skip! 
			continue
		}

		// sync blocks with peer
		err = self.syncBlocks(peer, status)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue // sync blocks failed
		}

		// merge peer_list with peer
		// status.KnownPeers[map]
		err = self.syncKnownPeers(status) 
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue// move to next on fail
		}

		// sync mempool txs
		if err := self.syncPendingTXs(peer, status.PendingTXs); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}
	}
}

func (self *Node) syncKnownPeers(
	// peer PeerNode,
	status StatusRes,
) error {
	for _, remotePeer := range status.KnownPeers {
		// check if we already have this node
		if !self.IsKnownPeer(remotePeer) {
			fmt.Printf("found new peer %s\n", remotePeer.TcpAddress())
			self.AddPeer(remotePeer)// add peer from other nodes
		}
	}

	return nil
}

// register with bootstrap node
func (self *Node) joinKnownPeers(peer PeerNode) error {

	// if already connected, we are registered
	// do not reconnect!!
	if peer.connected { return nil }

	// url to call and register
	// /node/peer
	var url = fmt.Sprintf(
		"http://%s%s?%s=%s&%s=%d&%s=%s",
		peer.TcpAddress(),
		endpointAddPeer,
		endpointAddPeerQueryKeyIP,
		self.info.IP,
		endpointAddPeerQueryKeyPort,
		self.info.Port,
		endpointAddPeerQueryKeyMiner,
		self.info.Account,
	)

	// call bootstrap /node/peer
	res, err := http.Get(url)
	if err != nil { return err }

	// response json to struct
	var addPeerRes = AddPeerRes{} 
	err = readRes(res, &addPeerRes)
	if err != nil { return err }

	// check for response error
	if addPeerRes.Error != "" {
		return fmt.Errorf("%s", addPeerRes.Error)
	}

	// we are register with this peer (bootstrap)
	// set our peer to connected
	knowPeer := self.knownPeers[peer.TcpAddress()]
	knowPeer.connected = addPeerRes.Success

	// overwrite peer inside map
	self.AddPeer(knowPeer)

	// check for response success
	if !addPeerRes.Success {
		return fmt.Errorf("unable to join peer '%s'", peer.TcpAddress())
	}

	return nil
}

// sync blocks with peers
func (self *Node) syncBlocks(
	peer PeerNode,
	status StatusRes,
) error {
	// block count
	var localBlockCount = self.state.GetChainLen() 	

	// if peer has zero/no blocks, return
	// latest block hash = [0 0 0 0 ...]
	if status.Hash.IsEmpty() {
		return nil
	}

	// if peer has less blocks than us, assume: they are behind
	if status.Number < localBlockCount {
		return nil
	}

	// check if both nodes have the genesis block
	// peer height = 1 && current_node.state.latestBlockHash != empty, meaning: already has genesisBlock or more
	if status.Number == 1 && !self.state.LatestBlockHash().IsEmpty() {
		fmt.Println("\n == i have genesis, thanks! == \n")
		return nil
	}

	// if local chain is behind?
	if localBlockCount < status.Number {
		var newBlocksCount = status.Number - localBlockCount
		fmt.Printf("\nfound %d new blocks from peer %s\n", newBlocksCount, peer.TcpAddress())

		// fetch blocks from peer
		// /node/sync?fromBlock=<our latest hash>
		blocks, err := fetchBlocksFromPeer(
			peer,
			// our own last block hash.
			self.state.LatestBlockHash(),
		)

		if err != nil { return err }

		// add new blocks from peer into state and block.db, but also send them through channel, so we stop mining
		if len(blocks) > 0 {
			for _, block := range blocks {
				_, err := self.state.AddBlock(block)
				if err != nil { return err }

				// send to miner, stop if your mining this block, it is already mined by others
				self.newSyncedBlocks <-block
			}
		}
	}
	return nil
}

// sync pending txs
func (self *Node) syncPendingTXs(
	peer PeerNode,
	txs []database.SignedTx,
) error {
	// basically: we go over our peer txs mempool and add each to our mempool, if have not seen it before
	for _, tx := range txs {
		err := self.AddPendingTX(tx, peer)
		if err != nil { return err }
	}
	return nil
}

// fetch: /node/status data
func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	// create a url
	// http://127.0.0.1:8081/node/status
	url := fmt.Sprintf("http://%s%s", peer.TcpAddress(), endpointStatus)

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

func fetchBlocksFromPeer(
	peer PeerNode,
	fromBlock database.Hash,
) ([]database.Block, error) {
	fmt.Printf("importing blocks from peer %s\n", peer.TcpAddress())

	// /node/sync?fromBlock
	var url = fmt.Sprintf(
		"http://%s%s?%s=%s",
		// address:ip of node, we get blocks from
		peer.TcpAddress(),
		endpointSync,
		endpointSyncQueryKeyFromBlock, // ?fromBlock
		fromBlock.Hex(),// hex of our lastBlockHash
	)

	// request blocks
	res, err := http.Get(url)
	if err != nil { return nil, err }

	var syncRes = SyncRes{}
	err = readRes(res, &syncRes)
	if err != nil { return nil, err }

	return syncRes.Blocks, nil
}