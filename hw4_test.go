package main

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestGetShardIDFromEveryNode(t *testing.T) {
	runContainers(2, 4)
	for port := range containersInfos {
		res := sendRequest(port, "GET", "shard/my_id", t, http.StatusOK, nil)
		if res["id"].(float64) >= float64(initNumOfShards) {
			t.Errorf("localhost:%v: shard_id must be < %v: %v", port, initNumOfShards, res["id"].(int))
		}
	}
}

func TestGetAllShardIDsFromEveryNode(t *testing.T) {
	runContainers(2, 4)
	maxShardID := initNumOfShards - 1
	for port := range containersInfos {
		res := sendRequest(port, "GET", "shard/all_ids", t, http.StatusOK, nil)
		_, ok := res["shard_ids"]
		if !ok {
			t.Errorf("localhost:%v: empty shard_ids received", port)
			t.FailNow()
		}
		var shardIDs []string
		shardIDs = strings.Split(res["shard_ids"].(string), ",")
		for shardID := range shardIDs {
			if maxShardID < shardID {
				t.Errorf("localhost:%v: shard_id must be <= %v: %v", port, maxShardID, shardID)
			}
			if res["result"].(string) != "Success" {
				t.Errorf("localhost:%v: result must be Success: %v", port, res["result"].(string))
			}
		}
	}
}

func TestMinNumMembersInShard(t *testing.T) {
	runContainers(2, 4)
	t.Run("Min", chechMinNumMembersInShardFromEveryNode)
}

func TestMinNumMembersInShardAfterAddingOneNode(t *testing.T) {
	runContainers(2, 4)
	newNetworkIP := runAContainer()
	putView := map[string]interface{}{"result": "Success", "msg": "Successfully added " + newNetworkIP + ":8080 to view"}
	checkView(t, "PUT", putView, newNetworkIP, !askEveryNode)
	chechMinNumMembersInShardFromEveryNode(t)
}

func TestKeyNotExist(t *testing.T) {
	runContainers(2, 4)
	expectedGetKey := map[string]interface{}{"result": "Error", "msg": "Key does not exist", "statusCode": http.StatusNotFound}
	checkGetKey(t, "bumba", expectedGetKey, askEveryNode)
}

func TestPutKeyFalse(t *testing.T) {
	runContainers(2, 4)
	expectedPutKey := map[string]interface{}{"replaced": false, "msg": "Added successfully", "statusCode": http.StatusCreated}
	checkPutKey(t, "Kuku", "Monkey", expectedPutKey, !askEveryNode)
}

func Test5Shard4Nodes(t *testing.T) {
	runContainers(1, 4)
	chechMinNumMembersInShardFromEveryNode(t)
	putShard := map[string]interface{}{"result": "Error", "msg": "Not enough nodes for 5 shards", "statusCode": http.StatusBadRequest}
	checkPutShard(t, 5, putShard)
}

func Test3Shard4Nodes(t *testing.T) {
	runContainers(1, 4)
	chechMinNumMembersInShardFromEveryNode(t)
	putShard := map[string]interface{}{"result": "Error", "msg": "Not enough nodes. 3 shards result in a nonfault tolerant shard", "statusCode": http.StatusBadRequest}
	checkPutShard(t, 3, putShard)
}

func Test2Shard4Nodes(t *testing.T) {
	runContainers(1, 4)
	chechMinNumMembersInShardFromEveryNode(t)
	putShard := map[string]interface{}{"result": "Success", "statusCode": http.StatusOK}
	checkPutShard(t, 2, putShard)

}

//--------------------------------------- Helper Functions----------------------------------/
func chechMinNumMembersInShardFromEveryNode(t *testing.T) {
	minNumMembers := int(math.Ceil(float64(initNumOfContainers) / float64(initNumOfShards)))
	for port := range containersInfos {
		for shardID := 0; shardID < initNumOfShards; shardID++ {
			res := sendRequest(port, "GET", "shard/members/"+strconv.Itoa(shardID), t, http.StatusOK, nil)
			_, ok := res["members"]
			if !ok {
				t.Errorf("localhost:%v: members must not be empty for shard_id %v", port, shardID)
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
				owner, isExist := res["owner"]
				if !isExist {
					t.Errorf("localhost:%v owner is not exist", port)
					t.FailNow()
				}
				ownerID = owner.(int)

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

func checkPutKey(t *testing.T, key string, val string, expected map[string]interface{}, putEveryone bool) {
	for port := range containersInfos {
		body := url.Values{}
		body.Add("val", val)
		res := sendRequest(port, "PUT", "keyValue-store/"+key, t, expected["statusCode"].(int), body)
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
				initNumOfShards = num
				maxShardID := initNumOfShards - 1
				var shardIDs []string
				shardIDs = strings.Split(res["shard_ids"].(string), ",")
				for shardID := range shardIDs {
					if maxShardID < shardID {
						t.Errorf("localhost:%v: shard_id must be <= %v: %v", port, maxShardID, shardID)
					}
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
