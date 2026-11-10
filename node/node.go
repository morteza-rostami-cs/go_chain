package node

import (
	"encoding/json"
	"fmt"
	database "go_chain/database"
	"io"
	"net/http"
)

// http port
const httpPort = 8081

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

// http server
func Run(dataDir string) error {
	// replay the blocks and set the state.Balances and also create /data files, if they do not exist
	state, err := database.NewStateFromDisk(dataDir)
	if err != nil { return err }

	defer state.Close()// close block.db file

	fmt.Printf("Listening on http port: %d\n", httpPort)

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

	// start the server
	return http.ListenAndServe(
		fmt.Sprintf(":%d", httpPort),// address listen on port :8080, on all available network interfaces
		nil)// nil means "use http.DefaultServeMux" — the global mux we just registered our two handlers into.
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

// sending err responses
func writeErrRes(res http.ResponseWriter, err error) {
	// struct to json, ignore err
	jsonErrRes, _ := json.Marshal(ErrRes{Error: err.Error()})

	// set response header, to json
	res.Header().Set("Content-Type", "application/json")
	// set http code
	res.WriteHeader(http.StatusInternalServerError)	// 500
	res.Write(jsonErrRes)// write json response body to client
}

// send successful response
func writeRes(
	res http.ResponseWriter, 
	content interface{}, //empty interface{}, cause data could be any types
){
	// struct to json
	contentJson, err := json.Marshal(content)
	if err != nil {
		writeErrRes(res, err)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)// 200
	res.Write(contentJson)
}

func readReq(
	req *http.Request,// request from client 
	reqBody interface{},// our struct to write data to
) error {
	// read request.body json -> into json []byte
	reqBodyJson, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf(
			"unable to read request body. %s", 
			err.Error(),
		)
	}

	defer req.Body.Close()// close request body

	// read body json bytes into our struct
	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal request body. %s\n", err.Error())
	}

	return nil// no error
}
