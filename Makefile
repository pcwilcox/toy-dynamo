# Makefile
# 
# CMPS 128, Fall 2018
#
# Lawrence Lawson     lelawson
# Pete Wilcox         pcwilcox
# Annie Shen          ashen7
# Victoria Tran       vilatran
#
# There's lots of stuff going on here. The primary targets are listed here:
#
#       make              - Builds the docker container
#
#       make all          - Run unit tests, then build the docker container,
#                           then run the integration tests
#
#       make run          - Build the docker container, then run a single instance
#
#       make cluster      - Build the docker container, then run 3 replicas
#
#       make clean        - This removes any running replicas
#
# When the docker container is build, a script copies all lines of this file which
# don't contain the string DELETE and writes them to a new file Makefile.docker. 
# The Dockerfile builds out of that file instead of this one. This is done because
# otherwise when Docker executes the build step, Make complains about the Docker
# command not being available (since our Docker base image doesn't have it).

# These build flags compile the binary specifically for Linux architecture
BUILD      = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
UNIT       = go test
DOCKERBLD  = docker build              # DELETE
DOCKERRUN  = docker run -d             # DELETE
DFLAGS     = -t ${CONTAINER} . 

# PIPENV manages dependencies for the Python test script
PIPENV     = pipenv run python3
TESTSCRIPT = hw4_test.py
TEST       = ${PIPENV} ${TESTSCRIPT}

# This is the container name
CONTAINER  = cs128-hw4

# This is the application name
EXEC       = app

# Add source files to this list
SOURCES    = main.go dbAccess.go app.go kvs.go restful.go values.go \
             gossip.go tcp.go rbtree.go hash.go shard.go forwardStructs.go docker_control.go

# Grabs the name of the current branch
BRANCH    := $(shell git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/\1/')

# Grabs the total number of commits
BUILDNUM  := $(shell git rev-list --count HEAD) 

# Grabs the last chunk of the current commit ID
HASH      := $(shell git rev-parse --short HEAD 2> /dev/null)

# Flags passed to the linker which define version strings in main.go
LDFLAGS1   = -X "main.branch=${BRANCH}" -X "main.hash=${HASH}" -X "main.build=${BUILDNUM}"

# These flags disable DWARF debugging info (used for GDB) and strips symbol table info
LDFLAGS2   = -w -s

# This just smashes the LDFLAGS together
LD         = -ldflags '${LDFLAGS1} ${LDFLAGS2}'

# These are defines for running containers based on our image
PORT       = 8080
PREFIX     = 10.0.0.
NET        = --net=mynet
IP         = --ip=${PREFIX}
VIEW       = -e VIEW="${PREFIX}2:${PORT},${PREFIX}3:${PORT},${PREFIX}4:${PORT},${PREFIX}5:${PORT},${PREFIX}6:${PORT},${PREFIX}7:${PORT}"
SHARDS     = -e S="3"
TAG        = ${CONTAINER}
NAME       = --name REPLICA_
REPLICA2   = ${NET} ${IP}2 ${VIEW} -e IP_PORT="${PREFIX}2:${PORT}" ${SHARDS} ${NAME}2 ${TAG} $(\n)
REPLICA3   = ${NET} ${IP}3 ${VIEW} -e IP_PORT="${PREFIX}3:${PORT}" ${SHARDS} ${NAME}3 ${TAG} $(\n)
REPLICA4   = ${NET} ${IP}4 ${VIEW} -e IP_PORT="${PREFIX}4:${PORT}" ${SHARDS} ${NAME}4 ${TAG} $(\n)
REPLICA5   = ${NET} ${IP}5 ${VIEW} -e IP_PORT="${PREFIX}5:${PORT}" ${SHARDS} ${NAME}5 ${TAG} $(\n)
REPLICA6   = ${NET} ${IP}6 ${VIEW} -e IP_PORT="${PREFIX}6:${PORT}" ${SHARDS} ${NAME}6 ${TAG} $(\n)
REPLICA7   = ${NET} ${IP}7 ${VIEW} -e IP_PORT="${PREFIX}7:${PORT}" ${SHARDS} ${NAME}7 ${TAG} $(\n)
REPLICAS   = ${REPLICA2} ${REPLICA3} ${REPLICA4} ${REPLICA5} ${REPLICA6} ${REPLICA7} 
SINGLE     = ${NET} ${IP}2 -e IP_PORT="${PREFIX}2:${PORT}" ${NAME}1 ${TAG}

# These three commands get a list of all containers running, all containers, and all images
RUNNING   := $(shell docker ps | grep REPLICA)                                  # DELETE
STOPPED   := $(shell docker ps -a | grep REPLICA)                               # DELETE
IMAGES    := $(shell docker images -a | grep REPLICA | awk ' $1 { print $1 } ') # DELETE

# This will run the docker build command, which builds the app
default : docker                            # DELETE

# Executes the build, unit tests, and runs the integration test
all : docker                                # DELETE
	${UNIT}                                 # DELETE
	${TEST}                                 # DELETE

# This runs 'go build ...' and is only executed by the Docker build process
app :
	${BUILD} -o ${EXEC} ${LD} ${SOURCES}

# This runs the unit tests
unit :
	${UNIT}

# This runs the integration tests
test : dockerclean docker                   # DELETE
	${TEST}                                 # DELETE

# This removes the binary
spotless : dockerclean                      # DELETE
ifneq (, $(shell ls | grep ${EXEC}))        # DELETE
	- rm ${EXEC}                            # DELETE
endif                                       # DELETE

# Alias for above
clean : spotless                            # DELETE

# Rebuild the thing
again :                                     # DELETE
	${MAKE} spotless all                    # DELETE

# This runs the Docker build command and also sets up the subnet
docker : network cleanfile                     # DELETE
	${DOCKERBLD} ${DFLAGS}                     # DELETE

# This makes a copy of the Makefile that doesn't try to run 
# Docker so we don't get a warning about it not being available.
cleanfile :
	awk '!/DELETE/' Makefile > Makefile.docker

# This builds the subnet in Docker
network :                                                   # DELETE
ifeq (, $(shell docker network ls | grep mynet))            # DELETE
	- sudo docker network create --subnet=10.0.0.0/16 mynet # DELETE
endif                                                       # DELETE

# This runs a single container detached
run : dockerclean docker                       # DELETE
	${DOCKERRUN} ${SINGLE}                     # DELETE

define \n


endef

# This builds a set of three nodes that should be able to talk to each other
cluster : dockerclean docker    # DELETE
	${DOCKERRUN} ${REPLICA2}    # DELETE
	${DOCKERRUN} ${REPLICA3}    # DELETE
	${DOCKERRUN} ${REPLICA4}    # DELETE
	${DOCKERRUN} ${REPLICA5}    # DELETE
	${DOCKERRUN} ${REPLICA6}    # DELETE
	${DOCKERRUN} ${REPLICA7}    # DELETE

dockerclean :                                                                                                # DELETE
	$(shell docker kill $$(docker ps -a | grep REPLICA | awk ' $$1 { print $$1 } ') > /dev/null 2>&1)        # DELETE
	$(shell docker rm $$(docker ps -a | grep REPLICA | awk ' $$1 { print $$1 } ') > /dev/null 2>&1)          # DELETE
	$(shell docker rmi $$(docker images -a | grep cs128 | awk ' $$1 { print $$1 } ') > /dev/null 2>&1)       # DELETE