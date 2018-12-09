import os
import sys
import requests
import time
import unittest
import json

import docker_control

import io

dockerBuildTag = "cs128-hw4"  # put the tag for your docker build here

hostIp = "localhost"

needSudo = False  # obviously if you need sudo, set this to True
# contact me imediately if setting this to True breaks things
# (I don't have a machine which needs sudo, so it has not been tested, although in theory it should be fine)

port_prefix = "808"

networkName = "mynet"  # the name of the network you created

# should be everything up to the last period of the subnet you specified when you
networkIpPrefix = "10.0.0."
# created your network

# sets number of seconds we sleep after certain actions to let data propagate through your system
propogationTime = 3
# you may lower this to speed up your testing if you know that your system is fast enough to propigate information faster than this
# I do not recomend increasing this

dc = docker_control.docker_controller(networkName, needSudo)


def getViewString(view):
    listOStrings = []
    for instance in view:
        listOStrings.append(instance["networkIpPortAddress"])

    return ",".join(listOStrings)


def viewMatch(collectedView, expectedView):
    collectedView = collectedView.split(",")
    expectedView = expectedView.split(",")

    if len(collectedView) != len(expectedView):
        return False

    for ipPort in expectedView:
        if ipPort in collectedView:
            collectedView.remove(ipPort)
        else:
            return False

    if len(collectedView) > 0:
        return False
    else:
        return True

# Basic Functionality
# These are the endpoints we should be able to hit
    # KVS Functions


def storeKeyValue(ipPort, key, value, payload):
    # print('PUT: http://%s/keyValue-store/%s'%(str(ipPort), key))
    return requests.put('http://%s/keyValue-store/%s' % (str(ipPort), key), data={'val': value, 'payload': payload})


def checkKey(ipPort, key, payload):
    # print('GET: http://%s/keyValue-store/search/%s'%(str(ipPort), key))
    return requests.get('http://%s/keyValue-store/search/%s' % (str(ipPort), key), data={'payload': payload})


def getKeyValue(ipPort, key, payload):
    # print('GET: http://%s/keyValue-store/%s'%(str(ipPort), key))
    return requests.get('http://%s/keyValue-store/%s' % (str(ipPort), key), data={'payload': payload})


def deleteKey(ipPort, key, payload):
    # print('DELETE: http://%s/keyValue-store/%s'%(str(ipPort), key))
    return requests.delete('http://%s/keyValue-store/%s' % (str(ipPort), key), data={'payload': payload})

    # Replication Functions


def addNode(ipPort, newAddress):
    # print('PUT: http://%s/view'%str(ipPort))
    return requests.put('http://%s/view' % str(ipPort), data={'ip_port': newAddress})


def removeNode(ipPort, oldAddress):
    # print('DELETE: http://%s/view'%str(ipPort))
    return requests.delete('http://%s/view' % str(ipPort), data={'ip_port': oldAddress})


def viewNetwork(ipPort):
    # print('GET: http://%s/view'%str(ipPort))
    return requests.get('http://%s/view' % str(ipPort))


def getShardId(ipPort):
    return requests.get('http://%s/shard/my_id' % str(ipPort))


def getAllShardIds(ipPort):
    return requests.get('http://%s/shard/all_ids' % str(ipPort))


def getMembers(ipPort, ID):
    return requests.get('http://%s/shard/members/%s' % (str(ipPort), str(ID)))


def getCount(ipPort, ID):
    return requests.get('http://%s/shard/count/%s' % (str(ipPort), str(ID)))


def changeShardNumber(ipPort, newNumber):
    return requests.put('http://%s/shard/changeShardNumber' % str(ipPort), data={'num': newNumber})

###########################################################################################


class TestHW4(unittest.TestCase):
    view = {}

    def setUp(self):
        self.view = dc.spinUpManyContainers(
            dockerBuildTag, hostIp, networkIpPrefix, port_prefix, 6, 3)

        for container in self.view:
            if " " in container["containerID"]:
                self.assertTrue(
                    False, "There is likely a problem in the settings of your ip addresses or network.")

        # dc.prepBlockade([instance["containerID"] for instance in self.view])

    def tearDown(self):
        dc.cleanUpDockerContainer()
        # dc.tearDownBlockade()

    def getPayload(self, ipPort, key):
        response = checkKey(ipPort, key, {})
        # print(response)
        data = response.json()
        return data["payload"]

    def partition(self, partitionList):
        truncatedList = []
        for partition in partitionList:
            truncatedPartition = []
            for node in partition:
                truncatedPartition.append(node[:12])
            truncatedPartition = ",".join(truncatedPartition)
            truncatedList.append(truncatedPartition)
        dc.partitionContainer(truncatedList)

    def partitionAll(self):
        listOLists = []
        for node in self.view:
            listOLists.append([node["containerID"]])
        self.partition(listOLists)

    def confirmAddKey(self, ipPort, key, value, expectedStatus, expectedMsg, expectedReplaced, payload={}):
        response = storeKeyValue(ipPort, key, value, payload)

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['msg'], expectedMsg)
        self.assertEqual(data['replaced'], expectedReplaced)

        return data["payload"]

    def confirmCheckKey(self, ipPort, key, expectedStatus, expectedResult, expectedIsExists, payload={}):
        response = checkKey(ipPort, key, payload)
        # print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['isExists'], expectedIsExists)

        return data["payload"]

    def confirmGetKey(self, ipPort, key, expectedStatus, expectedResult, expectedValue=None, expectedOwner=None, expectedMsg=None, payload={}):
        response = getKeyValue(ipPort, key, payload)
        # print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        if expectedValue != None and 'value' in data:
            self.assertEqual(data['value'], expectedValue)
        if expectedMsg != None and 'msg' in data:
            self.assertEqual(data['msg'], expectedMsg)
        if expectedOwner != None and 'owner' in data:
            self.assertEqual(data["owner"], expectedOwner)

        return data["payload"]

    def confirmDeleteKey(self, ipPort, key, expectedStatus, expectedResult, expectedMsg, payload={}):
        response = deleteKey(ipPort, key, payload)
        # print(response)

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['msg'], expectedMsg)

        return data["payload"]

    def confirmViewNetwork(self, ipPort, expectedStatus, expectedView):
        response = viewNetwork(ipPort)
        # print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()

        self.assertTrue(viewMatch(data['view'], expectedView), "%s != %s" % (
            data['view'], expectedView))

    def confirmAddNode(self, ipPort, newAddress, expectedStatus, expectedResult, expectedMsg):
        response = addNode(ipPort, newAddress)

        # print(response)

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['msg'], expectedMsg)

    def confirmDeleteNode(self, ipPort, removedAddress, expectedStatus, expectedResult, expectedMsg):
        response = removeNode(ipPort, removedAddress)
        # print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['msg'], expectedMsg)

    def checkGetMyShardId(self, ipPort, expectedStatus=200):
        response = getShardId(ipPort)

        self.assertEqual(response.status_code, expectedStatus)
        data = response.json()
        self.assertTrue('id' in data)

        return str(data['id'])

    def checkGetAllShardIds(self, ipPort, expectedStatus=200):
        response = getAllShardIds(ipPort)

        self.assertEqual(response.status_code, expectedStatus)
        data = response.json()
        return data["shard_ids"].split(",")

    def checkGetMembers(self, ipPort, ID, expectedStatus=200, expectedResult="Success", expectedMsg=None):
        response = getMembers(ipPort, ID)

        self.assertEqual(response.status_code, expectedStatus)
        data = response.json()

        self.assertEqual(data['result'], expectedResult)

        if "msg" in data and expectedMsg == None:
            self.assertEqual(data['msg'], expectedMsg)
        else:
            return data["members"].split(",")

    def getShardView(self, ipPort):
        shardIDs = self.checkGetAllShardIds(ipPort)
        shardView = {}
        for ID in shardIDs:
            shardView[ID] = self.checkGetMembers(ipPort, ID)
        return shardView

    def checkConsistentMembership(self, ipPort, ID):
        shard = self.checkGetMembers(ipPort, ID)
        for member in shard:
            ind = None
            for container in self.view:
                if container["networkIpPortAddress"] == member:
                    ind = container

            if ind != None:
                ip = ind["testScriptAddress"]
                self.assertEqual(self.checkGetMyShardId(ip), ID)

    ## Added by the student-wrote unit test ##
    def checkChangeShardNumber(self, ipPort, newNumber, expectedStatus, expectedResult, expectedShardIds, expectedMsg=""):
        response = changeShardNumber(ipPort, str(newNumber))

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)

        if expectedMsg:
            self.assertEqual(data['msg'], expectedMsg)
        else:
            self.assertEqual(data['shard_ids'], expectedShardIds)

    def checkGetCount(self, ipPort, ID, expectedStatus, expectedResult, expectedCount):
        response = getCount(ipPort, ID)

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['Count'], expectedCount)

    ##########################################################################
    ## Unit tests by Egan Gumiwang Pratama Bisma ##
    ##########################################################################

    # change S from 3 to 2 using changeShardNumber endpoint

    def test_f_decrease_shard(self):
        ipPort = self.view[0]["testScriptAddress"]
        targetNode = self.view[-1]["networkIpPortAddress"]

        initialShardIDs = self.checkGetAllShardIds(ipPort)

        self.checkChangeShardNumber(
            targetNode, 2, 200, "Success", "Abbey, Adelina")
        time.sleep(propogationTime)

        self.assertEqual(2, len(initialShardIDs)-1)

    # changing shard size from 1 to 2 should have 3 members in each shard
    def test_i_change_shard_size_from_one_to_two(self):
        self.test_h_change_shard_size_to_one()

        ipPortOne = self.view[0]["testScriptAddress"]
        ipPortTwo = self.view[1]["testScriptAddress"]

        members = self.checkGetMembers(ipPortOne, "Abbey")
        membersTwo = self.checkGetMembers(ipPortTwo, "Adelina")

        self.checkChangeShardNumber(
            ipPortOne, 2, 200, "Success", "Abbey, Adelina")

        membersOne = self.checkGetMembers(ipPortOne, 0)
        membersTwo = self.checkGetMembers(ipPortTwo, 0)

        self.assertEqual(3, len(membersOne))
        self.assertEqual(3, len(membersTwo))

    # when shard decreased and an isolated node moved to another shard,
    # its keys get shared to rest of shard members, and the key's owner is
    # consistent with new shard id
    def test_j_key_redistributed(self):
        ipPort = self.view[0]["testScriptAddress"]
        removedNode = self.view.pop()["networkIpPortAddress"]
        targetNode = self.view[-1]["networkIpPortAddress"]

        self.confirmAddKey(targetNode, 'key1', 'value1', 201,
                           "Added successfully", False, {})

        self.confirmDeleteNode(ipPort=ipPort,
                               removedAddress=removedNode,
                               expectedStatus=200,
                               expectedResult="Success",
                               expectedMsg="Successfully removed %s from view" % removedNode)

        time.sleep(propogationTime)

        # â—ï¸again, expected owner might be different based on different shard mechanic
        # we use regular hashing and sha1 as hash function -> hash(key1) % 2 = 1
        self.confirmGetKey(targetNode, 'key1', 200, "Success", 'value1', "1")

        self.confirmGetKey(ipPort, 'key1', 200, "Success", 'value1', "1")

    # setting shard to <=0 should be invalid
    def test_k_set_shard_to_zero(self):
        ipPort = self.view[0]["testScriptAddress"]
        self.checkChangeShardNumber(
            ipPort, 0, 400, "Error", "", "Must have at least one shard")


if __name__ == '__main__':
    unittest.main()
