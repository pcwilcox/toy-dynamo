/**********************************************
/ AUTHOR: Akobir Khamidov (akhamido@ucsc.edu) *
/ AUTHOR: Vien Van (vhvan@ucsc.edu)			  *
/ COPYRIGHT 2018 © by TEAMAWESOME			  *
***********************************************/
package main

import (
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var kvs map[string]string

//---------------------------------------------------- Messages -----------------------------------------------------------------------------------------------------------\
var goodDeleteNode = map[string]interface{}{"result": "Success", "statusCode": http.StatusOK}
var goodDeleteKey = map[string]interface{}{"result": "Success", "msg": "Key deleted", "statusCode": http.StatusOK}
var goodAddedKey = map[string]interface{}{"replaced": false, "msg": "Added successfully", "statusCode": http.StatusCreated}
var goodPutShard = map[string]interface{}{"result": "Success", "statusCode": http.StatusOK}
var notExistKey = map[string]interface{}{"result": "Error", "msg": "Key does not exist", "statusCode": http.StatusNotFound}
var crazyPutShard = map[string]interface{}{"result": "Error", "msg": "Not enough nodes for 5 shards", "statusCode": http.StatusBadRequest}
var faultyPutShard = map[string]interface{}{"result": "Error", "msg": "Not enough nodes. 3 shards result in a nonfault tolerant shard", "statusCode": http.StatusBadRequest}

//---------------------------------------------------- END ----------------------------------------------------------------------------------------------------------------\

//----------------------------------------------------SEARCH KEY -------------------------------------------------------\
// Write inside
//------------------------------------------------------ END -----------------------------------------------------------\
func TestShardIdIsUniqueFromEveryNode(t *testing.T) {
	runContainers(2, 4)
	key := "GuessWhat??"
	value := "Vien is Awesome!"
	initPutKey(t, key, value)

	var ownerIDs []float64
	for port := range containersInfos {
		res := sendRequest(port, "GET", "keyValue-store/"+key, t, http.StatusOK, nil)
		owner, _ := res["owner"]
		ownerID := owner.(float64)
		ownerIDs = append(ownerIDs, ownerID)
	}
	for i := 1; i < len(ownerIDs); i++ {
		if ownerIDs[i] != ownerIDs[0] {
			t.Errorf("Shard IDs are not unique!")
		}
	}
}

// 12 random keys stored
// Expecting: Sum of shard 1 and shard 2 should be equal to length of KVS
func TestGetAllCount(t *testing.T) {
	runContainers(2, 4)
	populateKVS(t, 12)
	time.Sleep(2 * time.Second)
	checkGetAllCount(t)
}

// 12 random keys stored
// 2 keys deleted
// Expected: total keys should be 10
func TestDeleteKeys(t *testing.T) {
	runContainers(2, 4)
	populateKVS(t, 12)
	checkGetAllCount(t)
	checkDeleteKey(t, goodDeleteKey, !askEveryNode)
	checkGetAllCount(t)
}

func TestGetShardIDFromEveryNode(t *testing.T) {
	runContainers(2, 4)
	for port := range containersInfos {
		res := sendRequest(port, "GET", "shard/my_id", t, http.StatusOK, nil)
		if res["id"].(float64) >= float64(initNumOfShards) {
			t.Errorf("localhost:%v: shard_id must be < %v: %v", port, initNumOfShards, res["id"].(int))
		}
	}
}

func TestKeyNotExist(t *testing.T) {
	runContainers(2, 4)
	checkGetKey(t, "bumba", notExistKey, askEveryNode)
}

func TestPutKeyFalse(t *testing.T) {
	runContainers(2, 4)
	checkPutKey(t, "Kuku", "Monkey", goodAddedKey, !askEveryNode)
}

// ------------------------------ Sharding - members ------------------------------------------/

// Expected: [A B C D] should not change to [A][B][C][D][]
func TestTryToIncrTo5Shard(t *testing.T) {
	runContainers(1, 4)
	chechMinNumMembersInShardFromEveryNode(t)
	checkPutShard(t, 5, crazyPutShard)
}

// Expected: [A B C D] should not change to [A][B][C D]
func TestTryToIncrTo3Shard(t *testing.T) {
	runContainers(1, 4)
	chechMinNumMembersInShardFromEveryNode(t)
	checkPutShard(t, 3, faultyPutShard)
}

// Expected: [A B C D] -> [A B][C D]
func TestIncrTo2Shards(t *testing.T) {
	runContainers(1, 4)
	chechMinNumMembersInShardFromEveryNode(t)
	checkPutShard(t, 2, goodPutShard)
	chechMinNumMembersInShardFromEveryNode(t)
}

// Expected: [A B] [C D] -> [A B C D]
func TestDecTo1Shard(t *testing.T) {
	runContainers(2, 4)
	chechMinNumMembersInShardFromEveryNode(t)
	checkPutShard(t, 1, goodPutShard)
	chechMinNumMembersInShardFromEveryNode(t)
}

// Expected: [A B C D ] [E F G] -> [A B C] [D E] [F G]
func TestIncrTo3Shard(t *testing.T) {
	runContainers(2, 7)
	chechMinNumMembersInShardFromEveryNode(t)
	checkPutShard(t, 3, goodPutShard)
	chechMinNumMembersInShardFromEveryNode(t)
}

// TODO it should check if total number of nodes in system to increase by X. len(containersInfos) == totalInSystem
func TestAddNode(t *testing.T) {
	// runContainers(2, 4)
	// newNetworkIP := runAContainer()
	// putView := map[string]interface{}{"result": "Success", "msg": "Successfully added " + newNetworkIP + ":8080 to view"}
	// checkView(t, "PUT", putView, newNetworkIP, !askEveryNode)
	// chechMinNumMembersInShardFromEveryNode(t)
}

// Expecting: stored keys stays same 12
func TestDeleteNode(t *testing.T) {
	runContainers(2, 4)
	populateKVS(t, 12)
	time.Sleep(2 * time.Second)
	checkGetAllCount(t)
	checkDeleteNode(t, goodDeleteNode, !askEveryNode)
	time.Sleep(2 * time.Second)
	checkGetAllCount(t)
}

// Expecting: 1 shard 1 node
func TestDeleteNode1Node1Shard(t *testing.T) {
	runContainers(1, 2)
	populateKVS(t, 12)
	time.Sleep(2 * time.Second)
	checkGetAllCount(t)
	checkDeleteNode(t, goodDeleteNode, !askEveryNode)
	time.Sleep(2 * time.Second)
	checkGetAllCount(t)
}

// ----------------------- Sharding - kvs Expected: size of key to stay same----------------------/
func TestIncreaseShardTo2from1(t *testing.T) {
	runContainers(1, 4)
	populateKVS(t, 23)
	checkPutShard(t, 2, goodPutShard)
	time.Sleep(2 * time.Second)
	checkGetAllCount(t)
}

func TestDecreaseShardTo3From4(t *testing.T) {
	runContainers(4, 8)
	populateKVS(t, 23)
	checkPutShard(t, 3, goodPutShard)
	time.Sleep(2 * time.Second)
	checkGetAllCount(t)
}

//--------------------------------------- Helper Functions----------------------------------/

func checkGetAllShardIDs(t *testing.T) []string {
	var shardIDs []string
	for port := range containersInfos {
		res := sendRequest(port, "GET", "shard/all_ids", t, http.StatusOK, nil)
		respShardIDs, ok := res["shard_ids"].(string)
		if !ok {
			t.Errorf("localhost:%v: empty shard_ids received", port)
			t.FailNow()
		}
		var expectedIDs string
		for i := 0; i < initNumOfShards; i++ {
			expectedIDs += strconv.Itoa(i) + ","
		}

		shardIDs = strings.Split(respShardIDs, ",")
		sort.Strings(shardIDs)
		expectedIDs = strings.TrimRight(expectedIDs, ",")
		if expectedIDs != strings.Join(shardIDs, ",") {
			t.Errorf("localhost:%v: Expected: %v: %v", port, expectedIDs, res["shard_ids"])
		}
		if res["result"].(string) != "Success" {
			t.Errorf("localhost:%v: result must be Success: %v", port, res["result"].(string))
		}
	}
	return shardIDs
}

func checkDeleteNode(t *testing.T, expected map[string]interface{}, deleteEveryNode bool) {
	for port := range containersInfos {
		body := url.Values{}
		deletePort, deleteNetworkIP := getPort()
		body.Add("ip_port", deleteNetworkIP)
		res := sendRequest(port, "DELETE", "view", t, http.StatusOK, body)

		result, isExist := res["result"].(string)
		if !isExist {
			t.Errorf("localhost:%v: Received empty response from deleteNode: %v", port, res)
			t.FailNow()
		}
		if result != expected["result"].(string) {
			t.Errorf("localhost:%v: result Expected: %v Received %v", port, expected["result"], result)
		}

		if result == "Success" {
			expectedMsg := "Successfully removed " + deleteNetworkIP + " from view"
			if res["msg"] != expectedMsg {
				t.Errorf("localhost:%v: msg Expected %v Received %v", port, expectedMsg, res["msg"])
				t.FailNow()
			}
			deleteNodeFromInfos(deletePort)
		} else {
			expectedMsg := deleteNetworkIP + " is not in current view"
			if res["msg"] != expectedMsg {
				t.Errorf("localhost:%v: msg Expected %v Received %v", port, expectedMsg, res["msg"])
				t.FailNow()
			}

		}
		if !deleteEveryNode {
			return
		}
	}
}

func chechMinNumMembersInShardFromEveryNode(t *testing.T) {
	minNumMembers := initNumOfContainers / initNumOfShards
	for port := range containersInfos {
		for shardID := 0; shardID < initNumOfShards; shardID++ {
			body := url.Values{}
			url := "shard/members/" + strconv.Itoa(shardID)
			res := sendRequest(port, "GET", url, t, http.StatusOK, body)
			_, ok := res["members"]
			if !ok {
				t.Errorf("localhost:%v: members must not be empty for shard_id %v: %v", port, shardID, res)
				t.FailNow()
			} else {
				members := strings.Split(res["members"].(string), ",")
				if len(members) < minNumMembers {
					t.Errorf("localhost:%v: not enough members in shard_id: %v. Members: %v", port, shardID, members)
				}
			}
		}
	}
}

func checkView(t *testing.T, typeReq string, expected map[string]interface{}, newNetworkIP string, askEveryone bool) {
	for port := range containersInfos {
		body := url.Values{}
		body.Add("ip_port", newNetworkIP+":8080")
		res := sendRequest(port, typeReq, "view", t, http.StatusOK, body)
		_, ok := res["msg"].(string)
		if !ok {
			t.Errorf("localhost:%v: unable to put view network ip:port: %v:8080", port, newNetworkIP)
			t.FailNow()
		} else {
			if res["msg"] != expected["msg"] {
				t.Errorf("localhost:%v: Expected: %v: %v", port, expected["msg"], res["msg"])
				t.FailNow()
			}
			if res["result"] != expected["result"] {
				t.Errorf("localhost:%v: Expected: %v : %v", port, expected["result"], res["result"])
				t.FailNow()
			}
		}
		if !askEveryone {
			break
		}
	}
}

func checkGetKey(t *testing.T, key string, expected map[string]interface{}, askEveryone bool) int {
	var ownerID = -1
	for port := range containersInfos {
		body := url.Values{}
		res := sendRequest(port, "GET", "keyValue-store/"+key, t, expected["statusCode"].(int), body)
		_, ok := res["result"].(string)
		if !ok {
			t.Errorf("localhost:%v: unable to get key: %v", port, key)
			t.FailNow()
		} else {
			if res["result"] != expected["result"] {
				t.Errorf("localhost:%v: Expected: %v : %v", port, expected["result"], res["result"])
				t.FailNow()
			}
			if expected["result"] == "Success" {
				if res["msg"] != expected["msg"] {
					t.Errorf("localhost:%v: Expected: %v: %v", port, expected["msg"], res["msg"])
					t.FailNow()
				}
				owner, isExist := res["owner"].(float64)
				if !isExist {
					t.Errorf("localhost:%v owner is not exist. Res: %v", port, res["owner"])
					t.FailNow()
				}
				ownerID = int(owner)

			} else if res["msg"] != expected["msg"] {
				t.Errorf("localhost:%v: Expected %v: %v", port, expected["msg"], res["msg"])
			}
		}
		if !askEveryone {
			break
		}
	}
	return ownerID
}

func checkDeleteKey(t *testing.T, expected map[string]interface{}, deleteFromEveryone bool) {
	for port := range containersInfos {
		body := url.Values{}
		randomKey := getRandomKey()
		res := sendRequest(port, "DELETE", "keyValue-store/"+randomKey, t, expected["statusCode"].(int), body)
		t.Logf("localhost:%v: deleted key: %v", port, randomKey)
		if res["result"].(string) != expected["result"].(string) {
			t.Errorf("localhost:%v: result Expected: %v: %v", port, expected["result"], res["result"])
		}
		msg, _ := res["msg"].(string)
		if msg != expected["msg"].(string) {
			t.Errorf("localhost:%v: msg Expected: %v: %v", port, expected["msg"], msg)
		}

		if expected["result"].(string) == "Success" {
			delete(kvs, randomKey)
		}
		if !deleteFromEveryone {
			return
		}
	}
}

func checkPutKey(t *testing.T, key string, val string, expected map[string]interface{}, putEveryone bool) {
	for port := range containersInfos {
		body := url.Values{}
		body.Add("val", val)
		res := sendRequest(port, "PUT", "keyValue-store/"+key, t, expected["statusCode"].(int), body)
		// t.Log(res)
		replaced, isExistReplaced := res["replaced"].(bool)
		msg, isExistMsg := res["msg"].(string)
		if !isExistMsg || !isExistReplaced {
			t.Errorf("localhost:%v: msg and replace not exist", port)
			t.FailNow()
		} else {
			if replaced != expected["replaced"].(bool) {
				t.Errorf("localhost:%v: Expected: %v: %v", port, expected["replaced"], replaced)
			}
			if msg != expected["msg"].(string) {
				t.Errorf("localhost:%v: Expected: %v: %v", port, expected["msg"], msg)
			}
		}
		if !putEveryone {
			return
		}
	}
}

func checkGetCount(t *testing.T, port string, shardID string, statusCode int, expected map[string]interface{}) int {
	var count float64

	body := url.Values{}
	res := sendRequest(port, "GET", "shard/count/"+shardID, t, statusCode, body)
	_, ok := res["result"].(string)
	if !ok {
		t.Errorf("localhost:%v: unable to get count from shard ID: %v", port, shardID)
		t.FailNow()
	} else {
		if res["result"] != expected["result"] {
			t.Errorf("Localhost:%v: Expected %v, got %v\n", port, expected["result"], res["result"])
			t.FailNow()
		}
		if res["result"] == "Error" {
			if res["msg"] != expected["msg"] {
				t.Errorf("Localhost:%v: Expected %v, got %v\n", port, expected["msg"], res["msg"])
				t.FailNow()
			}
		}
		localCount, isExist := res["Count"].(float64)
		if !isExist {
			t.Errorf("localhost:%v: unable to get count from shard ID: %v. ", port, shardID)
			t.FailNow()
		}
		count = localCount
	}

	return int(count)
}

func checkGetAllCount(t *testing.T) int {
	t.Logf("Checking stored key is equal to sum of shards")
	var allCount [][]int
	var totalCount int
	shardIDs := checkGetAllShardIDs(t)
	for _, shardID := range shardIDs {
		var count []int
		expectedGetKey := map[string]interface{}{"result": "Success", "statusCode": http.StatusOK}
		externalPorts := checkGetMembers(t, shardID, expectedGetKey)
		for _, port := range externalPorts {
			count = append(count, checkGetCount(t, "80"+port, shardID, http.StatusOK, expectedGetKey))
		}
		allCount = append(allCount, count)
	}
	for _, count := range allCount {
		isUnique(t, count)
		totalCount += count[0]
	}
	t.Logf("Received number of keys from shards: %v and sum of shards are %v", allCount, totalCount)
	if totalCount != len(kvs) {
		t.Errorf("Total Count %v != len(KVS) %v", totalCount, len(kvs))
		t.FailNow()
	}
	return totalCount
}

func checkGetMembers(t *testing.T, shardID string, expected map[string]interface{}) []string {
	var externalPorts []string
	for port := range containersInfos {
		body := url.Values{}
		res := sendRequest(port, "GET", "shard/members/"+shardID, t, expected["statusCode"].(int), body)
		if res["result"] != expected["result"].(string) {
			t.Errorf("localhost:%v: Expected result does not match actual result: %v != %v", port, expected["result"], res["result"])
			t.FailNow()
		}
		if res["result"] == "Error" {
			if res["msg"] != expected["msg"] {
				t.Errorf("localhost:%v: Expected msg does not match actual msg: %v != %v", port, expected["msg"], res["msg"])
				t.FailNow()
			}
		}
		// t.Logf("GET MEMBERS. Port: %v. Res: %v", port, res)
		members := strings.Split(res["members"].(string), ",")
		for _, member := range members {
			if !contains(externalPorts, getID(member)) {
				externalPorts = append(externalPorts, getID(member))
			}

		}
	}
	// t.Logf("externalPorts: %v", externalPorts)
	return externalPorts
}

func checkPutShard(t *testing.T, num int, expected map[string]interface{}) {
	for port := range containersInfos {
		body := url.Values{}
		body.Add("num", strconv.Itoa(num))
		res := sendRequest(port, "PUT", "shard/changeShardNumber", t, expected["statusCode"].(int), body)
		result, isExist := res["result"]
		if !isExist {
			t.Errorf("localhost:%v: received empty body: %v", port, res)
		} else {
			if result == "Success" {
				oldNumofShards := initNumOfShards
				initNumOfShards = num
				var expectedIDs string
				for i := 0; i < initNumOfShards; i++ {
					expectedIDs += strconv.Itoa(i) + ","
				}
				expectedIDs = strings.TrimRight(expectedIDs, ",")
				if expectedIDs != res["shard_ids"].(string) {
					t.Errorf("localhost:%v: Expected: %v: %v", port, expectedIDs, res["shard_ids"])
				} else {
					t.Logf("localhost:%v: Successfully changed number of shard %v -> %v", port, oldNumofShards, initNumOfShards)
				}
				return
			}
			if result != expected["result"] {
				t.Errorf("localhost:%v: Expected result: %v: %v", port, expected["result"], result)
			}
			if res["msg"] != expected["msg"] {
				t.Errorf("localhost:%v: Expected msg: %v: %v", port, expected["msg"], res["msg"])
			}
		}
		return
	}
}

// -------------------- HELPER HELPER functions ------------------------------\\\
func getRandomKey() string {
	for k := range kvs {
		return k
	}
	return "Empty kvs"
}
func getID(ipport string) string {
	s := strings.Split(ipport, ":")
	s = strings.Split(s[0], ".")

	id := s[3]
	return id
}

// Contains tells whether a contains x.
// Source: https://programming.guide/go/find-search-contains-slice.html
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func isUnique(t *testing.T, count []int) {
	for i := 1; i < len(count); i++ {
		if count[i] != count[0] {
			t.Errorf("Count from each nodes are not!")
			t.FailNow()
		}
	}
}

func initPutKey(t *testing.T, key string, value string) {
	expectedPutKey := map[string]interface{}{"replaced": false, "msg": "Added successfully", "statusCode": http.StatusCreated}
	checkPutKey(t, key, value, expectedPutKey, !askEveryNode)
}

func putKey(t *testing.T, port string, key string, val string, statusCode int) map[string]interface{} {
	body := url.Values{}
	body.Add("val", val)
	res := sendRequest(port, "PUT", "keyValue-store/"+key, t, statusCode, body)
	return res
}

func populateKVS(t *testing.T, numOfKeys int) {
	kvs = make(map[string]string)
	for i := 0; i < numOfKeys; i++ {
		key := "hello" + strconv.Itoa(i)
		val := "monkey"
		kvs[key] = val
		checkPutKey(t, key, val, goodAddedKey, !askEveryNode)
	}
	t.Logf("%v keys are created", len(kvs))
}

func deleteNodeFromInfos(port string) {
	delete(containersInfos, port)
	initNumOfContainers--
	if initNumOfShards/2 == 1 {
		initNumOfShards--
	}
}
