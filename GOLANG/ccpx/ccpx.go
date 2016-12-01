/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var pointIndexStr = "_pointindex"				//name for the key/value that will store a list of all known marbles
var transectionStr = "_tx"				//name for the key/value that will store all open trades
var tmpRelatedPoint = "_tmpRelatedPoint"
var tmpStr = "_tmpIndex"

var minimalTxStr = "_minimaltx"

type Point struct{
	Id string `json:"id"`					//the fieldtags are needed to keep case from bouncing around
	Owner string `json:"owner"`
}

type Description struct{
	Color string `json:"color"`
	Size int `json:"size"`
}

type Transaction struct{
	Id string `json:"txID"`					//user who created the open trade order
	Timestamp string `json:"timestamp"`			//utc timestamp of creation
	TraderA string  `json:"traderA"`				//description of desired marble
	TraderB string  `json:"traderB"`
	PointA string  `json:"pointA"`
	PointB string  `json:"pointB"`
	Related []Point `json:"related"`		//array of marbles willing to trade away
}

type AllTx struct{
	TXs []Transaction `json:"tx"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(pointIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	
	err = stub.PutState(tmpRelatedPoint, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	
	err = stub.PutState(tmpStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	var trades AllTx
	jsonAsBytes, _ = json.Marshal(trades)								//clear the open trade struct
	err = stub.PutState(transectionStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	err = stub.PutState(minimalTxStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "delete" {										//deletes an entity from its state
		res, err := t.Delete(stub, args)
		//cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_point" {									//create a new marble
		return t.init_point(stub, args)
	} else if function == "set_user" {										//change owner of a marble
		res, err := t.set_user(stub, args)
		//cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	} else if function == "findPointWithOwner"{
		res, err:= t.findPointWithOwner(stub, args)
		return res,err
	} else if function == "test"{
		return t.test(stub, args)
	} else if function == "init_transaction" {									//create a new trade order
		return t.init_transaction(stub, args)
	}
	/* 

	
	  else if function == "perform_trade" {									//forfill an open trade order
		res, err := t.perform_trade(stub, args)
		cleanTrades(stub)													//lets clean just in case
		return res, err
	} else if function == "remove_trade" {									//cancel an open trade order
		return t.remove_trade(stub, args)
	}*/
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}

	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var fcn, jsonResp string
	var err error


	fcn = args[0]
	if fcn == "read"{
		valAsbytes, err := stub.GetState(args[1])									//get the var from chaincode state
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + args[1] + "\"}"
			return nil, errors.New(jsonResp)
		}
		return valAsbytes, nil
	} else if fcn=="findLatest"{
		var seller = args[1]
		//var fetch = args[2]
		txAsbytes, err := stub.GetState(minimalTxStr)	
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + args[1] + "\"}"
			return nil, errors.New(jsonResp)
		}
		//some logic here
		var trans AllTx
		json.Unmarshal(txAsbytes, &trans)

		var processed AllTx

		for i := range trans.TXs{		
			if strings.Contains(trans.TXs[i].Id,seller){
				processed.TXs = append(processed.TXs,trans.TXs[i]);
			}
		}
		jsonAsBytes, _ := json.Marshal(processed)

		return jsonAsBytes, nil

	} else if fcn=="findRange"{
		/*var seller = args[1]
		var tx_from = args[2]
		var tx_to = args[3]
		txAsbytes, err := stub.GetState(minimalTxStr)
		//some logic here
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + args[1] + "\"}"
			return nil, errors.New(jsonResp)
		}
		return txAsbytes, nil*/
	}	
	return nil, err													//send it onward
}

/*func findPointWithOwner(stub shim.ChaincodeStubInterface, owner string )(m Point, err error){


}*/
// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	id := args[0]
	err := stub.DelState(id)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the marble index
	pointAsBytes, err := stub.GetState(pointIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get marble index")
	}
	var pointIndex []string
	json.Unmarshal(pointAsBytes, &pointIndex)								//un stringify it aka JSON.parse()
	
	//remove marble from index
	for i,val := range pointIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + id)
		if val == id{															//find the correct marble
			fmt.Println("found point")
			pointIndex = append(pointIndex[:i], pointIndex[i+1:]...)			//remove it
			for x:= range pointIndex{											//debug prints...
				fmt.Println(string(x) + " - " + pointIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(pointIndex)									//save new index
	err = stub.PutState(pointIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Point - create a new marble, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_point(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0        		1       
	// "SellerXhash", "Owner"
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	fmt.Println("- start init point")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}

	id := args[0]
	owner := strings.ToLower(args[1])
	

	//check if marble already exists
	pointAsBytes, err := stub.GetState(id)
	if err != nil {
		return nil, errors.New("Failed to get point id")
	}

	res := Point{}
	json.Unmarshal(pointAsBytes, &res)
	if res.Id == id{
		fmt.Println("This point arleady exists: " + id)
		fmt.Println(res);
		return nil, errors.New("This point arleady exists")				//all stop a marble by this name exists
	}
	
	//build the marble json string manually
	str := `{"id": "` + id + `", "owner": "` + owner + `"}`
	err = stub.PutState(id, []byte(str))									//store marble with id as key
	if err != nil {
		return nil, err
	}
		
	//get the marble index
	pointAsByte , err := stub.GetState(pointIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get marble index")
	}
	var pointIndex []string
	json.Unmarshal(pointAsByte, &pointIndex)							//un stringify it aka JSON.parse()
	
	//append
	pointIndex = append(pointIndex, id)									//add marble name to index list
	fmt.Println("! marble index: ", pointIndex)
	jsonAsBytes, _ := json.Marshal(pointIndex)
	err = stub.PutState(pointIndexStr, jsonAsBytes)						//store name of marble

	fmt.Println("- end init marble")
	return nil, nil
}
func (t *SimpleChaincode) init_transaction(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error	
	//	0        1      2     3      4      5       6
	//["bob", "blue", "16", "red", "16"] *"blue", "35*

	/*
	type Transaction struct{
	Id string `json:"txID"`					//user who created the open trade order
	Timestamp int64 `json:"timestamp"`			//utc timestamp of creation
	TraderA string  `json:"traderA"`				//description of desired marble
	TraderB string  `json:"traderB"`
	PointA string  `json:"pointA"`
	PointB string  `json:"pointB"`
	Related []Point `json:"related"`		//array of marbles willing to trade away
}
	*/
	if len(args) < 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 4?")
	}

	open := Transaction{}
	open.Id = args[0]
	open.Timestamp = time.Now().String()
	open.TraderA = args[1]
	open.TraderB = args[2]
	open.PointA = args[3]
	open.PointB = args[4]
	
	fmt.Println("- start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	//get the open trade struct
	tradesAsBytes, err := stub.GetState(minimalTxStr)
	if err != nil {
		return nil, errors.New("Failed to get TXs")
	}
	var trades AllTx
	json.Unmarshal(tradesAsBytes, &trades)										//un stringify it aka JSON.parse()
	
	trades.TXs = append(trades.TXs, open);						//append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(trades)
	err = stub.PutState(minimalTxStr, jsonAsBytes)								//rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("- end open trade")
	return nil, nil
}

// ============================================================================================================================
// Set User Permission on Point
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	pointAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Point{}
	json.Unmarshal(pointAsBytes, &res)										//un stringify it aka JSON.parse()
	res.Owner = args[1]														//change the user
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the marble with id as key
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end set user")
	return nil, nil
}

func (t *SimpleChaincode) test(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	fmt.Println("- start test fcn")
	fmt.Println(args[0] + " - " + args[1])

	//get the open trade struct
	tmpAsBytes, err := stub.GetState(tmpStr)
	if err != nil {
		return nil, errors.New("Failed to get TXs")
	}
	var tmps []string

	json.Unmarshal(tmpAsBytes, &tmps)	
	
	return nil, nil
}

// ============================================================================================================================
// Open Trade - create an open trade for a marble you want with marbles you have 
// ============================================================================================================================
/*func (t *SimpleChaincode) open_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var will_size int
	var trade_away Description
	
	//	0        1      2     3      4      5       6
	//["bob", "blue", "16", "red", "16"] *"blue", "35*
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 5?")
	}
	if len(args)%2 == 0{
		return nil, errors.New("Incorrect number of arguments. Expecting an odd number")
	}

	size1, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	open := Transaction{}
	open.User = args[0]
	open.Timestamp = makeTimestamp()											//use timestamp as an ID
	open.Want.Color = args[1]
	open.Want.Size =  size1
	fmt.Println("- start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	for i:=3; i < len(args); i++ {												//create and append each willing trade
		will_size, err = strconv.Atoi(args[i + 1])
		if err != nil {
			msg := "is not a numeric string " + args[i + 1]
			fmt.Println(msg)
			return nil, errors.New(msg)
		}
		
		trade_away = Description{}
		trade_away.Color = args[i]
		trade_away.Size =  will_size
		fmt.Println("! created trade_away: " + args[i])
		jsonAsBytes, _ = json.Marshal(trade_away)
		err = stub.PutState("_debug2", jsonAsBytes)
		
		open.Willing = append(open.Willing, trade_away)
		fmt.Println("! appended willing to open")
		i++;
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(transectionStr)
	if err != nil {
		return nil, errors.New("Failed to get TXs")
	}
	var trades AllTx
	json.Unmarshal(tradesAsBytes, &trades)										//un stringify it aka JSON.parse()
	
	trades.TXs = append(trades.TXs, open);						//append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(trades)
	err = stub.PutState(transectionStr, jsonAsBytes)								//rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("- end open trade")
	return nil, nil
}*/

// ============================================================================================================================
// Perform Trade - close an open trade and move ownership
// ============================================================================================================================
/*
func (t *SimpleChaincode) perform_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//	0		1					2					3				4					5
	//[data.id, data.closer.user, data.closer.name, data.opener.user, data.opener.color, data.opener.size]
	if len(args) < 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}
	
	fmt.Println("- start close trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}
	
	size, err := strconv.Atoi(args[5])
	if err != nil {
		return nil, errors.New("6th argument must be a numeric string")
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(transectionStr)
	if err != nil {
		return nil, errors.New("Failed to get TXs")
	}
	var trades AllTx
	json.Unmarshal(tradesAsBytes, &trades)															//un stringify it aka JSON.parse()
	
	for i := range trades.TXs{																//look for the trade
		fmt.Println("looking at " + strconv.FormatInt(trades.TXs[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.TXs[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			
			
			pointAsBytes, err := stub.GetState(args[2])
			if err != nil {
				return nil, errors.New("Failed to get thing")
			}
			closersMarble := Point{}
			json.Unmarshal(pointAsBytes, &closersMarble)											//un stringify it aka JSON.parse()
			
			//verify if marble meets trade requirements
			if closersMarble.Color != trades.TXs[i].Want.Color || closersMarble.Size != trades.TXs[i].Want.Size {
				msg := "marble in input does not meet trade requriements"
				fmt.Println(msg)
				return nil, errors.New(msg)
			}
			
			marble, e := findMarble4Trade(stub, trades.TXs[i].User, args[4], size)			//find a marble that is suitable from opener
			if(e == nil){
				fmt.Println("! no errors, proceeding")

				t.set_user(stub, []string{args[2], trades.TXs[i].User})						//change owner of selected marble, closer -> opener
				t.set_user(stub, []string{marble.Name, args[1]})									//change owner of selected marble, opener -> closer
			
				trades.TXs = append(trades.TXs[:i], trades.TXs[i+1:]...)		//remove trade
				jsonAsBytes, _ := json.Marshal(trades)
				err = stub.PutState(transectionStr, jsonAsBytes)										//rewrite open orders
				if err != nil {
					return nil, err
				}
			}
		}
	}
	fmt.Println("- end close trade")
	return nil, nil
}*/

// ============================================================================================================================
// findMarble4Trade - look for a matching marble that this user owns and return it
// ============================================================================================================================

func (t *SimpleChaincode) findPointWithOwner(stub shim.ChaincodeStubInterface, args []string )([]byte, error){
//func findPointWithOwner(stub shim.ChaincodeStubInterface, owner string )(m Point, err error){
	var fail []byte
	var success []byte

	success = []byte("success")

	var owner = args[0]
	fmt.Println("- start find marble 4 trade")
	fmt.Println("looking for " + owner)

	//get the marble index
	pointAsBytes, err := stub.GetState(pointIndexStr)
	if err != nil {
		return fail, errors.New("Failed to get marble index")
	}
	var pointIndex []string
	json.Unmarshal(pointAsBytes, &pointIndex)								//un stringify it aka JSON.parse()

	var pointRelated []string

	for i:= range pointIndex{													//iter through all the marbles
		//fmt.Println("looking @ marble name: " + pointIndex[i]);

		pointAsBytes, err := stub.GetState(pointIndex[i])						//grab this marble
		if err != nil {
			return fail, errors.New("Failed to get marble")
		}
		res := Point{}
		json.Unmarshal(pointAsBytes, &res)										//un stringify it aka JSON.parse()
		//fmt.Println("looking @ " + res.User + ", " + res.Color + ", " + strconv.Itoa(res.Size));
		
		//check for user && color && size
		if strings.ToLower(res.Owner) == strings.ToLower(owner){
			//get the marble index
			pointAsByte , err := stub.GetState(tmpRelatedPoint) //gettmpindex
			if err != nil {
				return nil, errors.New("Failed to get marble index")
			}
			json.Unmarshal(pointAsByte, &tmpRelatedPoint)							//un stringify it aka JSON.parse()
			
			//append
			pointIndex = append(pointIndex, res.Id)									//add marble name to index list
			jsonAsBytes, _ := json.Marshal(pointIndex)
			err = stub.PutState(tmpRelatedPoint, jsonAsBytes)						//store name of marble
		}

	}
	if pointRelated == nil {
		return success,nil
	}
	//fmt.Println("- end find marble 4 trade - error")
	return fail, errors.New("Did not find marble to use in this trade")
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

// ============================================================================================================================
// Remove Open Trade - close an open trade
// ============================================================================================================================
/*
func (t *SimpleChaincode) remove_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//	0
	//[data.id]
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	fmt.Println("- start remove trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(transectionStr)
	if err != nil {
		return nil, errors.New("Failed to get TXs")
	}
	var trades AllTx
	json.Unmarshal(tradesAsBytes, &trades)																//un stringify it aka JSON.parse()
	
	for i := range trades.TXs{																	//look for the trade
		//fmt.Println("looking at " + strconv.FormatInt(trades.TXs[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.TXs[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			trades.TXs = append(trades.TXs[:i], trades.TXs[i+1:]...)				//remove this trade
			jsonAsBytes, _ := json.Marshal(trades)
			err = stub.PutState(transectionStr, jsonAsBytes)												//rewrite open orders
			if err != nil {
				return nil, err
			}
			break
		}
	}
	
	fmt.Println("- end remove trade")
	return nil, nil
}*/

// ============================================================================================================================
// Clean Up Open Trades - make sure open trades are still possible, remove choices that are no longer possible, remove trades that have no valid choices
// ============================================================================================================================
/*
func cleanTrades(stub shim.ChaincodeStubInterface)(err error){
	var didWork = false
	fmt.Println("- start clean trades")
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(transectionStr)
	if err != nil {
		return errors.New("Failed to get TXs")
	}
	var trades AllTx
	json.Unmarshal(tradesAsBytes, &trades)																		//un stringify it aka JSON.parse()
	
	fmt.Println("# trades " + strconv.Itoa(len(trades.TXs)))
	for i:=0; i<len(trades.TXs); {																		//iter over all the known open trades
		fmt.Println(strconv.Itoa(i) + ": looking at trade " + strconv.FormatInt(trades.TXs[i].Timestamp, 10))
		
		fmt.Println("# options " + strconv.Itoa(len(trades.TXs[i].Willing)))
		for x:=0; x<len(trades.TXs[i].Willing); {														//find a marble that is suitable
			fmt.Println("! on next option " + strconv.Itoa(i) + ":" + strconv.Itoa(x))
			_, e := findMarble4Trade(stub, trades.TXs[i].User, trades.TXs[i].Willing[x].Color, trades.TXs[i].Willing[x].Size)
			if(e != nil){
				fmt.Println("! errors with this option, removing option")
				didWork = true
				trades.TXs[i].Willing = append(trades.TXs[i].Willing[:x], trades.TXs[i].Willing[x+1:]...)	//remove this option
				x--;
			}else{
				fmt.Println("! this option is fine")
			}
			
			x++
			fmt.Println("! x:" + strconv.Itoa(x))
			if x >= len(trades.TXs[i].Willing) {														//things might have shifted, recalcuate
				break
			}
		}
		
		if len(trades.TXs[i].Willing) == 0 {
			fmt.Println("! no more options for this trade, removing trade")
			didWork = true
			trades.TXs = append(trades.TXs[:i], trades.TXs[i+1:]...)					//remove this trade
			i--;
		}
		
		i++
		fmt.Println("! i:" + strconv.Itoa(i))
		if i >= len(trades.TXs) {																	//things might have shifted, recalcuate
			break
		}
	}

	if(didWork){
		fmt.Println("! saving open trade changes")
		jsonAsBytes, _ := json.Marshal(trades)
		err = stub.PutState(transectionStr, jsonAsBytes)														//rewrite open orders
		if err != nil {
			return err
		}
	}else{
		fmt.Println("! all open trades are fine")
	}

	fmt.Println("- end clean trades")
	return nil
}*/