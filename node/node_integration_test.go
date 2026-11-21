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
	var morAcc = database.NewAccount(config.TestKsMortezaAccount)
	var babaAcc = database.NewAccount(config.TestKsBabaYagaAccount)

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
	// if err != nil { t.Fatal(err) }

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
	// function to shutdown node.Run()
	ctx, closeNodeFunc := context.WithTimeout(context.Background(), time.Minute*15)

	// a go that creates a transaction
	go func() {
		// sleep 1 sec => so our node has time to start up
		time.Sleep(time.Second * 1)

		var tx = database.NewTx(
			morAcc,
			babaAcc,
			1,
			1,//nonce
			"",
		)

		// sign tx
		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx, 
			morAcc,// account to sign with 
			config.TestKsAccountsPwd, // password
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
			2,// nonce
			"",
		) 

		// sign tx
		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx, 
			morAcc,// account to sign with 
			config.TestKsAccountsPwd, // password
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
	var morAcc = database.NewAccount(config.TestKsMortezaAccount)
	var babaAcc = database.NewAccount(config.TestKsBabaYagaAccount)

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
		1,
		"",
	)
	// current time
	var tx2 = database.NewTx(
		morAcc,
		babaAcc,
		2,
		2,// nonce
		"",
	)

	// sign txs
	signedTx1, err := wallet.SignTxWithKeystoreAccount(
		tx1, 
		morAcc, 
		config.TestKsAccountsPwd, 
		wallet.GetKeystoreDirPath(dataDir),
	)
	if err != nil {
		t.Error(err)
		// return
	}

	signedTx2, err := wallet.SignTxWithKeystoreAccount(
		tx2, 
		morAcc, 
		config.TestKsAccountsPwd, 
		wallet.GetKeystoreDirPath(dataDir),
	)
	if err != nil {
		t.Error(err)
		// return
	}

	// hash second tx2
	tx2Hash, _ := signedTx2.Hash()

	var pBlock = NewPendingBlock(
		database.Hash{},
		1,// block_number
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

// attack:: hacker sends a valid sig + different tx
func TestNode_ForgedTx(t *testing.T) {
	// setup test env
	dataDir, morAcc, babaAcc, err := setupTestNodeDir(1000_000)
	if err != nil { t.Error(err) }

	defer fs.RemoveDir(dataDir)// remove the test data dir at the end

	var node = New(
		dataDir, 
		DefaultIP, 
		DefaultHttpPort, 
		morAcc, 
		PeerNode{},
	)

	// 15 min timeout for node
	ctx, _ := context.WithTimeout(context.Background(), time.Minute * 15)

	// peer node who signed tx
	morPeerNode := NewPeerNode(
		DefaultIP, 
		DefaultHttpPort, 
		false,// not bootstrap 
		morAcc, // miner account
		true,// connected to our current node
	)

	// tx
	var txValue uint = 5
	var txNonce uint = 1

	var tx = database.NewTx(morAcc, babaAcc, txValue, txNonce, "")

	// sign tx
	validSignedTx, err := wallet.SignTxWithKeystoreAccount(
		tx,
		morAcc, // signer
		config.TestKsAccountsPwd,
		wallet.GetKeystoreDirPath(dataDir),
	)
	
	if err != nil {
		t.Error(err)
		// return
	}

	// push to mempool and trigger mining
	go func() {
		time.Sleep(time.Second * 1)
		_ = node.AddPendingTX(validSignedTx, morPeerNode)
	}()

	go func() {
		// wait for mining process of first legit tx to start.
		time.Sleep(time.Second * (miningIntervalSeconds + 1))

		// forge a tx, using the sig of previous tx
		var forgedTx = database.NewTx(morAcc, babaAcc, txValue, txNonce, "")

		var forgedSignedTx = database.NewSignedTx(
			forgedTx,
			validSignedTx.Sig,
		)

		// we add the forged one to the pool, before archived happening at the end of mining! after the first of is mined => this one will be rejected, cause: has the same nonce!
		// although:: this is a different tx, so it has a different hash to it passes the archived test!
		_ = node.AddPendingTX(forgedSignedTx, morPeerNode)
	}()

	// node, has to wait 10 min => cause: the second block never gets mined => is invalid and errors out of miner!! this test has to wait!!
	_ = node.Run(ctx)

	if node.state.CurrentBlockNumber() != 1 {
		t.Fatal("mined two blocks! second tx is forged.")
	}

	if node.state.Balances[babaAcc] != txValue {
		t.Fatal("forged tx succeeded.") 
	}
}

// attack:: hacker re-submit the same tx twice
func TestNode_ReplayedTx(t *testing.T) {
	// setup test env
	dataDir, morAcc, babaAcc, err := setupTestNodeDir(1000_000)
	if err != nil { t.Error(err) }

	defer fs.RemoveDir(dataDir)// remove the test data dir at the end

	var node = New(
		dataDir, 
		DefaultIP, 
		DefaultHttpPort, 
		morAcc, 
		PeerNode{},
	)

	// 15 min timeout for node
	ctx, closeNode := context.WithCancel(context.Background())

	// peer node who signed tx
	morPeerNode := NewPeerNode(
		DefaultIP, 
		DefaultHttpPort, 
		false,// not bootstrap 
		morAcc, // miner account
		true,// connected to our current node
	)

	// sender of forged tx
	babaPeerNode := NewPeerNode(
		DefaultIP, 
		8082, 
		false,// not bootstrap 
		babaAcc, // miner account
		true,// connected to our current node
	)

	// tx
	var txValue uint = 5
	var txNonce uint = 1

	var tx = database.NewTx(morAcc, babaAcc, txValue, txNonce, "")

	// sign tx
	validSignedTx, err := wallet.SignTxWithKeystoreAccount(
		tx,
		morAcc, // signer
		config.TestKsAccountsPwd,
		wallet.GetKeystoreDirPath(dataDir),
	)

	if err != nil {
		t.Error(err)
		// return
	}

	// push to mempool and trigger mining
	_ = node.AddPendingTX(validSignedTx, morPeerNode)

	go func() {
		// starts in 7 sec => sooner that miner
		ticker := time.NewTicker(time.Second * (miningIntervalSeconds - 3))

		var wasReplayedTxAdded = false

		for range ticker.C {
			// after the first block is mined, we attack with the same tx (twice) for replay!!
			// we need to wait so, block is added, so nonce = 1
			if node.state.CurrentBlockNumber() == 1 {

				// check: we already added the same tx twice into our mempool and the mining is over
				if wasReplayedTxAdded && !node.isMining {
					closeNode()
					return
				}

				// clear archive, and we can replay the same tx => this is like a server crash!
				clear(node.archivedTXs)

				// submit the same tx again
				_ = node.AddPendingTX(
					validSignedTx,
					babaPeerNode,
				)
				wasReplayedTxAdded = true

				// if second block is mined?
				if node.state.CurrentBlockNumber() == 2 {
					closeNode()
					return
				}
			}
		}
	}()

	_ = node.Run(ctx)

	if node.state.CurrentBlockNumber() == 2 {
		t.Errorf("mined two blocks! second tx is forged.")
		return
	}

	if node.state.Balances[babaAcc] == txValue * 2 {
		t.Errorf("replay attack was successful, second tx was counted.")
	}
}

// test: if you spam with multiple txs => you have to pay txFee
// so: you do not do that that shit, now that we are applying txFee to each tx
func TestNode_MiningSpamTransactions(t *testing.T){
	// balances, from, to, miner
	var morBalance = uint(1000)
	var babaBalance = uint(0)
	var minerBalance = uint(0)

	// get keys => key.address = eth miner address
	minerKey, err := wallet.NewRandomKey()
	if err != nil { t.Fatal(err) }

	var miner = minerKey.Address// eth address

	// dataDir, account 1, account2, err
	dataDir, morAcc, babaAcc, err := setupTestNodeDir(morBalance)
	if err != nil { t.Fatal(err) }

	// close test folder
	defer fs.RemoveDir(dataDir)

	// make a node
	var node = New(
		dataDir, 
		DefaultIP, 
		DefaultHttpPort, 
		miner, // real account, gets gas 
		PeerNode{},//we do not send thing between nodes
	)
	ctx, closeNode := context.WithCancel(context.Background())

	// peer who add txs
	var minerPeerNode = NewPeerNode(DefaultIP, DefaultHttpPort, false, miner, true)

	// transaction value
	var txValue = uint(200)

	// max tx count to send
	var txCount = uint(4)

	// sign and add 4 tx to the pool
	// i = uint => cause: nonce has to be uint
	for i := uint(1); i <= txCount; i++ {
		time.Sleep(time.Second)
		// 1, 2, 3 => each tx has a different nonce
		var txNonce = i
		var tx = database.NewTx(morAcc, babaAcc, txValue, txNonce, "")

		signedTx, err := wallet.SignTxWithKeystoreAccount(tx, morAcc, config.TestKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil { t.Fatal(err) } 

		// add SignedTx
		_ = node.AddPendingTX(signedTx, minerPeerNode)
	}

	// so, we add txs at the same time, so they should be mined together
	// now: we just have to check for any mined block, every 10 sec
	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for range ticker.C {
			if !node.state.LatestBlockHash().IsEmpty() {
				closeNode()
				return
			}
		}
	}()

	_ = node.Run(ctx)

	// check expected balances for mor, baba and miner

	// 4 tx => 4 times gasFee payment
	expectedMorBalance := morBalance - (txCount * txValue) - (txCount * database.TxFee)

	// baba only gets the real value
	expectedBabaBalance := babaBalance + (txCount * txValue)

	// miner gets the gasFee + block reward
	expectedMinerBalance := minerBalance + database.BlockReward + (txCount * database.TxFee)

	if node.state.Balances[morAcc] != expectedMorBalance {
		t.Errorf(
			"mor balance is not correct. expected %d and got %d",
			expectedMorBalance,
			node.state.Balances[morAcc],
		)
		return
	}

	// check babayaga for receiving tx.value
	if node.state.Balances[babaAcc] != expectedBabaBalance {
		t.Errorf(
			"baba balance is wrong. expected %d, got %d",
			expectedBabaBalance,
			node.state.Balances[babaAcc],
		)
	} 

	// check miner balance, after mining 4 txs
	if node.state.Balances[miner] != expectedMinerBalance {
		t.Errorf(
			"miner balance is wrong. expected %d, got %d",
			expectedMinerBalance,
			node.state.Balances[miner],
		)
		return
	}

	t.Logf("mor final balance: %d TMC", node.state.Balances[morAcc])
	t.Logf("baba final balance: %d TMC", node.state.Balances[babaAcc])
	t.Logf("miner final balance: %d TMC", node.state.Balances[miner])
}


func getTestDataDirPath() (string) {
	// just put data in some temp folder 
	return filepath.Join(os.TempDir(), "tmc_test")
}

// for test, we have are keystore files hard coded!! because we want fixed account_addresses, not just some random keys each time!
// so, we load our own files => into /data dir, for keystore files!
func copyKeystoreFilesIntoTestDataDirPath(dataDir string) error {
	// open morteza keystore file
	mortezaSrcKs, err := os.Open(config.TestKsMortezaFile) 

	if err != nil { return err }
	defer mortezaSrcKs.Close()

	// set the dir for keystore
	// /test/keystore/sdfsd_sddssd
	var ksDir = filepath.Join(wallet.GetKeystoreDirPath(dataDir))

	// create path and set folder permission
	err = os.Mkdir(ksDir, 0777)
	if err != nil { return err }

	// create morteza keystore file
	mortezaDstKs, err := os.Create(filepath.Join(ksDir, config.TestKsMortezaFile)) 
	if err != nil {
		return err
	}

	defer mortezaDstKs.Close()

	// copy json content into the file
	_, err = io.Copy(mortezaDstKs, mortezaSrcKs)
	if err != nil { return err }

	// repeat the same thing for babayaga file

	babayagaSrcKs, err := os.Open(config.TestKsBabaYagaFile) 
	if err != nil { return err }

	defer babayagaSrcKs.Close()

	babayagaDstKs, err := os.Create(filepath.Join(ksDir, config.TestKsBabaYagaFile))
	if err != nil { return err }

	defer babayagaDstKs.Close()

	_, err = io.Copy(babayagaDstKs, babayagaSrcKs)
	if err != nil { return err }

	return nil
}

// get dataDir, accounts for test , and init things for test
// when using this: clean up fs.RemoveDir()
// send balance for testing gas fee
func setupTestNodeDir(morBalance uint) (
	dataDir string,
	morAcc database.Account,
	babaAcc database.Account,
	err error,
) {

	// tmc_test
	dataDir = getTestDataDirPath()

	// clear the shit test files => so we can create a new one for each test
	fmt.Println("removing test dir! in setupTestNodeDir()")
	fmt.Println(dataDir)

	err = fs.RemoveDir(dataDir)
	if err != nil {
		return "", database.Account{}, database.Account{}, err
	}

	morAcc = database.NewAccount(config.TestKsMortezaAccount)
	babaAcc = database.NewAccount(config.TestKsBabaYagaAccount)

	// setup genesis.json file
	var balances = make(map[database.Account]uint)
	// balances[morAcc] = 1000_000
	balances[morAcc] = morBalance
	var genesis = database.Genesis{Balances: balances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		return "", database.Account{}, database.Account{}, err
	}

	// init the data folders
	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	if err != nil {
		return "", database.Account{}, database.Account{}, err
	}

	// copy the test keystore files => to /data
	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		return "", database.Account{}, database.Account{}, err
	}

	return dataDir, morAcc, babaAcc, nil
}
