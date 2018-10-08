import unittest
import subprocess
import requests
import os

# This is set in the test script, default 8080
# Need to strip a lot of quotes because the scripts store it as '"5000"' or as "'5000'"
PORT = int(os.environ['PORT_EXT'].strip("'").strip('"'))


class TestHW1(unittest.TestCase):

    def test1(self):
        res = requests.get('http://localhost:'+str(PORT)+'/test')
        self.assertEqual(res.text, 'GET request received',
                         msg='Incorrect response to GET request to /test endpoint')
        self.assertEqual(
            res.status_code, 200, msg='Did not return status 200 to GET request to /test endpoint')

    def test2(self):
        res = requests.post('http://localhost:'+str(PORT) +
                            '/test?msg=ACoolMessage')
        self.assertEqual(res.text, 'POST message received: ACoolMessage',
                         msg='Incorrect response to POST request to /test endpoint')
        self.assertEqual(
            res.status_code, 200, msg='Did not return status 200 to POST request to /test endpoint')

    def test3(self):
        res = requests.get('http://localhost:'+str(PORT)+'/hello')
        self.assertEqual(res.text, 'Hello world!',
                         msg='Incorrect response to /hello endpoint')


if __name__ == '__main__':
    unittest.main()
