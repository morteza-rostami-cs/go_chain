package node

import (
	"context"
	"encoding/json"
	"fmt"
	"go_chain/config"
	"go_chain/database"
	"go_chain/fs"
	"go_chain/wallet"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// basically: we want to have control over the eth_addresses stored inside keystore files
// fixed account_address => that we store in test keystore file
const testKsMortezaAccount = "0x3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
const testKsBabaYagaAccount = "0x6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"

// each eth_address => get a keystore file => with the following files
const testKsMortezaFile = "test_morteza--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
const testKsBabaYagaFile = "test_babayaga--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
// test keystore password
const testKsAccountsPwd = "security123"

// test node running and shutdown
func TestNode_Run(t *testing.T) {

	dataDir := getTestDataDirPath()
	// remove the old test folder
	err := fs.RemoveDir(dataDir)
	if err != nil { t.Fatal(err) }

	// run a Node
	var node = New(
		dataDir, 
		DefaultIP, 
		DefaultHttpPort, 
		database.NewAccount(DefaultMiner), 
		PeerNode{},
	)

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
	// test accounts addresses
	var morAcc = database.NewAccount(testKsMortezaAccount)
	var babaAcc = database.NewAccount(testKsBabaYagaAccount)

	//set our genesis.json content for test
	var genesisBalances = make(map[database.Account]uint) 
	genesisBalances[morAcc] = 1000_000
	// genesis.json content
	var genesis = database.Genesis{Balances: genesisBalances}

	// struct to json
	genesisJson, err := json.Marshal(genesis)
	if err != nil { t.Fatal(err) }
	
	// test /data folder
	var dataDir = getTestDataDirPath() 
	// err = fs.RemoveDir(dataDir)
	if err != nil { t.Fatal(err) }

	// initialize data folders and write genesis.json and so on
	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	if err != nil { t.Fatal(err) }

	// remove test /data folder at the end
	defer fs.RemoveDir(dataDir)

	// copy test json files into /data/keystore
	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		t.Fatal(err)
	}

	var info = NewPeerNode(
		DefaultIP,
		DefaultHttpPort,
		true,
		database.NewAccount(""),
		false,
	)

	var node = New(
		dataDir, 
		DefaultIP, 
		DefaultHttpPort, 
		database.NewAccount(DefaultMiner), 
		info,
	)

	// a 15 minutes timeout for our node, enough for some mining!
	// 		function to shutdown node.Run()
	ctx, closeNodeFunc := context.WithTimeout(context.Background(), time.Minute*15)

	// a go that creates a transaction
	go func() {
		// sleep 1 sec => so our node has time to start up
		time.Sleep(time.Second * 1)

		var tx = database.NewTx(
			morAcc,
			babaAcc,
			1,
			"",
		)

		// sign tx
		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx, 
			morAcc,// account to sign with 
			testKsAccountsPwd, // password
			wallet.GetKeystoreDirPath(dataDir),
		)
		if err != nil {
			t.Error(err)
			return
		}

		_ = node.AddPendingTX(signedTx, info)
	}()

	// create the second transaction
	go func() {
		// wait 30 sec
		time.Sleep(time.Second * 20)

		var tx = database.NewTx(
			database.NewAccount(config.MortezaAccount),
			database.NewAccount(config.BabaYagaAccount),
			2,
			"",
		)		

		// sign tx
		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx, 
			morAcc,// account to sign with 
			testKsAccountsPwd, // password
			wallet.GetKeystoreDirPath(dataDir),
		)
		if err != nil {
			t.Error(err)
			return
		}

		_ = node.AddPendingTX(signedTx, info)
	}()

	// each 10 sec wait for node.state latest blockNumber, if 2 !! means that we successfully mined 2 blocks => close node_server
	go func() {
		ticker := time.NewTicker(10 * time.Second)

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

	//accounts
	var morAcc = database.NewAccount(testKsMortezaAccount)
	var babaAcc = database.NewAccount(testKsBabaYagaAccount)

	var dataDir = getTestDataDirPath()
	// err := fs.RemoveDir(dataDir)
	// if err != nil { t.Fatal(err) }

	
	// genesis data
	var balances = make(map[database.Account]uint)
	balances[morAcc] = 1000_000
	var genesis = database.Genesis{Balances: balances}

	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}

	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	defer fs.RemoveDir(dataDir)

	if err != nil {
		t.Fatal(err)
	}

	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil { t.Fatal(err) }

	// channel for errors coming from goroutines
	errCh := make(chan error, 1)
	// wait for balance check
	var balanceDone = make(chan struct{})

	// peer node => the node that mines block sooner
	var info = NewPeerNode(
		DefaultIP,
		DefaultHttpPort,
		false,
		database.NewAccount(""),// zero account
		true,
	)

	var node = New(
		dataDir,
		DefaultIP, 
		8082, 
		babaAcc,
		// we do not send mined block over http
		PeerNode{},// empty bootstrap
	)

	// context for node shutdown, 15 minutes
	ctx, closeNode := context.WithCancel(
		context.Background(), 
		// time.Minute * 15,
	)

	/*
	{"from":wallet.MortezaAccount,"to":"babayaga","value":1,"data":"","time":wallet.BabaYagaAccount248871086}
	*/ 

	var tx1 = database.NewTx(
		morAcc,
		babaAcc,
		1,
		"",
	)
	// current time
	var tx2 = database.NewTx(
		morAcc,
		babaAcc,
		2,
		"",
	)

	// sign txs
	signedTx1, err := wallet.SignTxWithKeystoreAccount(
		tx1, 
		morAcc, 
		testKsAccountsPwd, 
		wallet.GetKeystoreDirPath(dataDir),
	)
	if err != nil {
		t.Error(err)
		return
	}

	signedTx2, err := wallet.SignTxWithKeystoreAccount(
		tx2, 
		morAcc, 
		testKsAccountsPwd, 
		wallet.GetKeystoreDirPath(dataDir),
	)
	if err != nil {
		t.Error(err)
		return
	}

	// hash second tx2
	tx2Hash, _ := signedTx2.Hash()

	var pBlock = NewPendingBlock(
		database.Hash{},
		0,// block_number
		morAcc, // miner
		[]database.SignedTx{signedTx1},
	)

	// morAcc mines a block
	validSyncedBlock, err := Mine(ctx, pBlock)
	if err != nil {
		t.Fatal(err)
	}

	// morteza:: adds two txs to the mempool
	go func() {
		// sleep 8 sec
		time.Sleep(time.Second * (miningIntervalSeconds - 2))

		// this one, our node start mining, but then we signal => it is already mined, so, stop!
		err := node.AddPendingTX(signedTx1, info)
		if err != nil { 
			errCh <-err
			return	
		}

		// this one it mines itself
		err = node.AddPendingTX(signedTx2, info)
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
		fmt.Println("\nabout the add a new Block.")
		_, err := node.state.AddBlock(validSyncedBlock)
		if err != nil {
			fmt.Println(err)
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
		if len(node.pendingTXs) != 1 || !onlyTx2Pending {
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
			if node.state.GetChainLen() == 2 {
				fmt.Println("\n==============")
				fmt.Println("two blocks mined. we are done!")
				fmt.Println("\n==============")

				closeNode()
				return
			}			
		}
	}()

	// another go, to check state.balances, to know who got paid the block reward
	go func() {
		time.Sleep(time.Second * 2)
		// close channel at the end
		defer close(balanceDone)

		// get the starting balances
		// currentNode is babayaga, and morteza mines the block
		var startMortezaBalance = node.state.Balances[morAcc] 
		var startBabayagaBalance = node.state.Balances[babaAcc]

		// wait for the timeout signal (after mining)
		<-ctx.Done()

		// get the balances after mining
		var endMortezaBalance = node.state.Balances[morAcc]
		var endBabayagaBalance = node.state.Balances[babaAcc]

		// expected balances => each have mined one block
		var expectedMortezaBalance = startMortezaBalance - tx1.Value - tx2.Value + database.BlockReward
		var expectedBabayagaBalance = startBabayagaBalance + tx1.Value + tx2.Value + database.BlockReward

		if endMortezaBalance != expectedMortezaBalance {
			errCh <-fmt.Errorf("morteza expected end balance is %d not %d", expectedMortezaBalance, endMortezaBalance)
			return
		}

		if endBabayagaBalance != expectedBabayagaBalance {
			errCh <-fmt.Errorf("BabaYaga expected end balance is %d not %d", expectedBabayagaBalance, endBabayagaBalance)
			return
		}

		t.Logf("Starting morteza balance: %d", startMortezaBalance)
		t.Logf("Starting BabaYaga balance: %d", startBabayagaBalance)
		t.Logf("Ending morteza balance: %d", endMortezaBalance)
		t.Logf("Ending BabaYaga balance: %d", endBabayagaBalance)
	}()

	// run node
	_ = node.Run(ctx)

   // wait for balance check
	<-balanceDone

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

func getTestDataDirPath() (string) {
	// just put data in some temp folder 
	return filepath.Join(os.TempDir(), "tmc_test")
}

// for test, we have are keystore files hard coded!! because we want fixed account_addresses, not just some random keys each time!
// so, we load our own files => into /data dir, for keystore files!
func copyKeystoreFilesIntoTestDataDirPath(dataDir string) error {
	// open morteza keystore file
	mortezaSrcKs, err := os.Open(testKsMortezaFile) 
	if err != nil { return err }

	defer mortezaSrcKs.Close()

	// set the dir for keystore
	// /test/keystore/sdfsd_sddssd
	var ksDir = filepath.Join(wallet.GetKeystoreDirPath(dataDir))

	// create path and set folder permission
	err = os.Mkdir(ksDir, 0777)
	if err != nil { return err } 

	// create morteza keystore file
	mortezaDstKs, err := os.Create(filepath.Join(ksDir, testKsMortezaFile)) 
	if err != nil {
		return err
	}

	defer mortezaDstKs.Close()

	// copy json content into the file
	_, err = io.Copy(mortezaDstKs, mortezaSrcKs)
	if err != nil { return err }

	// repeat the same thing for babayaga file

	babayagaSrcKs, err := os.Open(testKsBabaYagaFile) 
	if err != nil { return err }

	defer babayagaSrcKs.Close()

	babayagaDstKs, err := os.Create(filepath.Join(ksDir, testKsBabaYagaFile))
	if err != nil { return err }

	defer babayagaDstKs.Close()

	_, err = io.Copy(babayagaDstKs, babayagaSrcKs)
	if err != nil { return err }

	return nil
}