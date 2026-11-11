package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
