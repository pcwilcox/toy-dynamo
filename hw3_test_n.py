# last update: 11/19 - cleaned up the last several tests to not try to delete the
#                      addresses the host would use for comunication, but rather
#                      an address they would use to communicate between each other.
# past updates:
# 11/17/18 - changed to use subnets, since Mac and Linux apparently really need them
# 11/10/18 - fixed the expected result of GET view

import os
import sys
import requests
import time
import unittest
import json

import docker_control

dockerBuildTag = "cs128-hw3"  # put the tag for your docker build here

hostIp = "localhost"  # this can be localhost again

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
    print('PUT: http://%s/keyValue-store/%s' % (str(ipPort), key))
    return requests.put('http://%s/keyValue-store/%s' % (str(ipPort), key), data={'val': value, 'payload': json.dumps(payload)}, timeout=5)


def checkKey(ipPort, key, payload):
    print('GET: http://%s/keyValue-store/search/%s' % (str(ipPort), key))
    return requests.get('http://%s/keyValue-store/search/%s' % (str(ipPort), key), data={'payload': json.dumps(payload)})


def getKeyValue(ipPort, key, payload):
    print('GET: http://%s/keyValue-store/%s' % (str(ipPort), key))
    return requests.get('http://%s/keyValue-store/%s' % (str(ipPort), key), data={'payload': json.dumps(payload)})


def deleteKey(ipPort, key, payload):
    print('DELETE: http://%s/keyValue-store/%s' % (str(ipPort), key))
    return requests.delete('http://%s/keyValue-store/%s' % (str(ipPort), key), data={'payload': json.dumps(payload)})

    # Replication Functions


def addNode(ipPort, newAddress):
    print('PUT: http://%s/view' % str(ipPort))
    return requests.put('http://%s/view' % str(ipPort), data={'ip_port': newAddress})


def removeNode(ipPort, oldAddress):
    print('DELETE: http://%s/view' % str(ipPort))
    return requests.delete('http://%s/view' % str(ipPort), data={'ip_port': oldAddress})


def viewNetwork(ipPort):
    print('GET: http://%s/view' % str(ipPort))
    return requests.get('http://%s/view' % str(ipPort))

###########################################################################################


class TestHW3(unittest.TestCase):
    view = {}

    def setUp(self):
        self.view = dc.spinUpManyContainers(
            dockerBuildTag, hostIp, networkIpPrefix, port_prefix, 2)

        for container in self.view:
            if " " in container["containerID"]:
                self.assertTrue(
                    False, "There is likely a problem in the settings of your ip addresses or network.")

    def tearDown(self):
        dc.cleanUpDockerContainer()

    def getPayload(self, ipPort, key):
        response = checkKey(ipPort, key, {})
        print(response)
        data = response.json()
        return data["payload"]

    def confirmAddKey(self, ipPort, key, value, expectedStatus, expectedMsg, expectedReplaced, payload={}):
        response = storeKeyValue(ipPort, key, value, payload)

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['msg'], expectedMsg)
        self.assertEqual(data['replaced'], expectedReplaced)

        return data["payload"]

    def confirmCheckKey(self, ipPort, key, expectedStatus, expectedResult, expectedIsExists, payload={}):
        response = checkKey(ipPort, key, payload)
        print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['isExists'], expectedIsExists)

        return data["payload"]

    def confirmGetKey(self, ipPort, key, expectedStatus, expectedResult, expectedValue=None, expectedMsg=None, payload={}):
        response = getKeyValue(ipPort, key, payload)
        print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        if expectedValue != None and 'value' in data:
            self.assertEqual(data['value'], expectedValue)
        if expectedMsg != None and 'msg' in data:
            self.assertEqual(data['msg'], expectedMsg)

        return data["payload"]

    def confirmDeleteKey(self, ipPort, key, expectedStatus, expectedResult, expectedMsg, payload={}):
        response = deleteKey(ipPort, key, payload)
        print(response)

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['msg'], expectedMsg)

        return data["payload"]

    def confirmViewNetwork(self, ipPort, expectedStatus, expectedView):
        response = viewNetwork(ipPort)
        print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()

        self.assertTrue(viewMatch(data['view'], expectedView), "%s != %s" % (
            data['view'], expectedView))

    def confirmAddNode(self, ipPort, newAddress, expectedStatus, expectedResult, expectedMsg):
        response = addNode(ipPort, newAddress)

        print(response)

        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['msg'], expectedMsg)

    def confirmDeleteNode(self, ipPort, removedAddress, expectedStatus, expectedResult, expectedMsg):
        response = removeNode(ipPort, removedAddress)
        print(response)
        self.assertEqual(response.status_code, expectedStatus)

        data = response.json()
        self.assertEqual(data['result'], expectedResult)
        self.assertEqual(data['msg'], expectedMsg)

    def test_n_replication_remove_node(self):
        key = "HeyWhereDidYouGo"
        value = "IllHoldYourStuffWhileYoureGone"

        stationaryNode = self.view[0]["testScriptAddress"]
        removedNode = self.view.pop()

        payload = self.getPayload(removedNode["testScriptAddress"], key)

        payload = self.confirmAddKey(ipPort=removedNode["testScriptAddress"],
                                     key=key,
                                     value=value,
                                     expectedStatus=200,
                                     expectedMsg="Added successfully",
                                     expectedReplaced=False,
                                     payload=payload)

        time.sleep(propogationTime)

        self.confirmDeleteNode(ipPort=stationaryNode,
                               removedAddress=removedNode["networkIpPortAddress"],
                               expectedStatus=200,
                               expectedResult="Success",
                               expectedMsg="Successfully removed %s from view" % removedNode["networkIpPortAddress"])

        payload = self.confirmCheckKey(ipPort=stationaryNode,
                                       key=key,
                                       expectedStatus=200,
                                       expectedResult="Success",
                                       expectedIsExists=True,
                                       payload=payload)

        payload = self.confirmGetKey(ipPort=stationaryNode,
                                     key=key,
                                     expectedStatus=200,
                                     expectedResult="Success",
                                     expectedValue=value,
                                     payload=payload)


if __name__ == '__main__':
    unittest.main()
