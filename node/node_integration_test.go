package node

import (
	"context"
	"fmt"
	"go_chain/database"
	"go_chain/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// test node running and shutdown
func TestNode_Run(t *testing.T) {

	dataDir := getTestDataDirPath()
	// remove the old test folder
	err := fs.RemoveDir(dataDir)
	if err != nil { t.Fatal(err) }

	// run a Node
	var node = New(dataDir, DefaultIP, DefaultHttpPort, PeerNode{})

	// a time out context, after 5 sec ctx.Done() channel closes
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	// run node, in .Run() we close server if ctx.Done(), which will trigger after 5 sec
	err = node.Run(ctx)
	if err.Error() != "http: Server closed" {
		t.Fatal("node server should close after 5sec")
	}
}

// test node mining feature
func TestNode_Mining(t *testing.T) {
	var dataDir = getTestDataDirPath() 
	err := fs.RemoveDir(dataDir)
	if err != nil { t.Fatal(err) }

	var node = New(dataDir, DefaultIP, DefaultHttpPort, PeerNode{})

	// a 15 minutes timeout for our node, enough for some mining!
	// 		function to shutdown node.Run()
	ctx, closeNodeFunc := context.WithTimeout(context.Background(), time.Minute*15)

	// a go that creates a transaction
	go func() {
		// sleep 1 sec => so our node has time to start up
		time.Sleep(time.Second * 1)

		var tx = database.NewTx(
			database.NewAccount("morteza"),
			database.NewAccount("babayaga"),
			1,
			"",
		)

		// same 8081 node, adds a tx to it's mempool
		// so node 8082 => add a transaction to its mempool
		var fromPeer = NewPeerNode(DefaultIP, DefaultHttpPort, false, true)
		_ = node.AddPendingTX(tx, fromPeer)
	}()

	// create the second transaction
	go func() {
		// wait 30 sec
		time.Sleep(time.Second * 30)
		var tx = database.NewTx(
			database.Account("morteza"),
			database.Account("babayaga"),
			2,
			"",
		)		
		var fromPeer = NewPeerNode(DefaultIP, DefaultHttpPort, false, true)
		_ = node.AddPendingTX(tx, fromPeer)
	}()

	// each 10 sec wait for node.state latest blockNumber, if 2 !! means that we successfully mined 2 blocks => close node_server
	go func() {
		ticker := time.NewTicker(10 * time.Second)

		// for {
		// 	select{
		// 	case <-ticker.C:
		// 		if node.state.GetChainLen() == 2 {
		// 			fmt.Println("2 blocks mined! goodbye!")
		// 			closeNodeFunc()// close node server
		// 			return
		// 		}
		// 	}
		// }
		for range ticker.C {
			if node.state.GetChainLen() == 2 {
				fmt.Println("2 block mined. goodbye!")
				closeNodeFunc()// close node server
				return
			} else {
				fmt.Println("still mining", node.state.GetChainLen())	
			}
		}
	}()

	// run our node
	_ = node.Run(ctx)// http server blocks

	// out of http server => check if we mined 2 blocks
	if node.state.GetChainLen() != 2 {
		t.Fatal("node did not mine 2 blocks! fuck!")
	}
}

// test if miner stops, on other peer's success
func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	var dataDir = getTestDataDirPath()
	err := fs.RemoveDir(dataDir)
	if err != nil { t.Fatal(err) }

	// channel for errors coming from goroutines
	errCh := make(chan error, 1)

	var node = New(dataDir, DefaultIP, DefaultHttpPort, PeerNode{})
	// context for node shutdown, 15 minutes
	ctx, closeNode := context.WithCancel(
		context.Background(), 
		// time.Minute * 15,
	)

	var tx = database.Tx{
		From: database.NewAccount("morteza"),
		To: database.NewAccount("babayaga"),
		Value: 1,
		// fixed time, so we can define a nonce, that passes mining rule
		Time: 1579451695,
		Data: "",
	}

	// current time
	var tx2 = database.NewTx(
		database.NewAccount("morteza"),
		database.NewAccount("babayaga"),
		2,
		"",
	)
	tx2Hash, _ := tx2.Hash()

	// an already mined block
	var validSyncedBlock = database.NewBlock(
		database.Hash{},
		1,
		12612040,// nonce
		1579451704,// time
		[]database.Tx{tx},
	)

	// add two txs to the mempool
	go func() {
		// sleep 8 sec
		time.Sleep(time.Second * (miningIntervalSeconds - 2))

		var fromPeer = NewPeerNode(DefaultIP, DefaultHttpPort, false, true) 

		// this one, our node start mining, but then we signal => it is already mined, so, stop!
		err := node.AddPendingTX(tx, fromPeer)
		if err != nil { 
			errCh <-err
			return	
		}

		// this one it mines itself
		err = node.AddPendingTX(tx2, fromPeer)
		if err != nil { 
			errCh <-err
			return
		}
	}()

	// after a while we signal the mined blocked, to miner
	go func() {
		// 12 sec, let it start mining!
		time.Sleep(time.Second * (miningIntervalSeconds + 2))

		if !node.isMining {
			errCh <- fmt.Errorf("node should be mining tx1. fuck.")
			return
		}

		// add the mined block to block.db
		_, err := node.state.AddBlock(validSyncedBlock)
		if err != nil {
			errCh <- fmt.Errorf("%w", err)
			return
		}
		fmt.Println("mined block was added, by other peer.")

		// send mined node through channel, to stop the other miner
		node.newSyncedBlocks <-validSyncedBlock
		time.Sleep(time.Second * 2) // wait 2 sec for cancellation

		// check: miner should have shutdown now
		if node.isMining {
			// t.Fatal("new received block, should have cancelled miner!")
			errCh <-fmt.Errorf("block is received! miner should stop! but is running.")
			return
		}
		
		// now only tx2, must be in mempool
		_, onlyTx2Pending := node.pendingTXs[tx2Hash.Hex()]
		// only one tx in pool, and it is tx2
		if len(node.pendingTXs) != 1 && !onlyTx2Pending {
			// t.Fatal("something is fucked! tx1 is still in pool!")
			errCh <- fmt.Errorf("fuck__ tx1 is still in pool")
			return
		}
	
		// wait for mining to start for tx2
		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !node.isMining {
			// t.Fatal("should be mining tx2 now")
			errCh <- fmt.Errorf("why node is not mining, tx2.")
			return
		}
	}()

	// check for 2 blocks mined every 2 second
	go func() {
		ticker := time.NewTicker(10 * time.Second)

		defer ticker.Stop()

		for range ticker.C {
			fmt.Println("====")
			if node.state.LatestBlock() != nil {
				fmt.Println(node.state.LatestBlock().Header.Number)
			} else {
				fmt.Println("node latest block is nil")
				fmt.Println(node.state)
			}
			fmt.Println(node.state.GetChainLen())
			if node.state.GetChainLen() == 2 {
				fmt.Println("\n==============")
				fmt.Println("two blocks mined. we are done!")
				fmt.Println("\n==============")

				closeNode()
				return
			}			
		}
	}()


	// run node
	_ = node.Run(ctx)
	if node.state.GetChainLen() != 2 {
		t.Fatal("something wrong.nigga! we did not mined 2 blocks!")
	}
	if len(node.pendingTXs) != 0 {
		t.Fatal("fuck nigga! we have more than zero txs in the pool.")
	}

	// handle the error out of goroutines
	// just check the error at the end, after node.Run stop
	select{
	case err := <-errCh:
		t.Fatal(err)
	default:// so we do not block
	}
}

func getTestDataDirPath() string {
	// just put data in some temp folder 
	return filepath.Join(os.TempDir(), ".tmc_test")
}