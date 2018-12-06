# CMPS128 Distributed Systems

Authors:

 * Lawrence Lawson - lelawson@ucsc.edu
 * Pete Wilcox     - pcwilcox@ucsc.edu
 * Annie Shen      - ashen7@ucsc.edu
 * Victoria Tran   - vilatran@ucsc.edu

This is our team project for CMPS128 Fall 2018. We've developed it using Go and Docker. ~~It runs on a Jenkins server in Pete's apartment for CI testing.~~ Jenkins is terrible so we're going to implement CircleCI.

To execute, clone the repo run `make docker` to build the container, `make cluster` to launch a trio of containers, `make test` to run the latest integration test.
