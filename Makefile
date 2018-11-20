# Makefile
# 
# CMPS 128, Fall 2018
#
# Lawrence Lawson     lelawson
# Pete Wilcox         pcwilcox
# Annie Shen          ashen7
# Victoria Tran       vilatran
#
# Incorporates compile-time variables into the build

# Predefined stuff
BUILD      = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
UNIT       = go test
DOCKERBLD  = docker build                                                                                   # DELETE
DOCKERRUN  = docker run -d                                                                                  # DELETE
DFLAGS     = -t ${CONTAINER} . 
PIPENV     = pipenv run python3
TESTSCRIPT = hw3_test.py
TEST       = ${PIPENV} ${TESTSCRIPT}

# This is the container name
CONTAINER  = cs128-hw3

# This is the application name
EXEC       = app

# Add source files to this list
SOURCES    = main.go dbAccess.go app.go kvs.go restful.go values.go view.go gossip.go tcp.go

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

# These are defines for running a cluster
PORT       = 8080
PREFIX     = 10.0.0.
NET        = --net=mynet
IP         = --ip=${PREFIX}
VIEW       = -e VIEW="${PREFIX}2:${PORT},${PREFIX}3:${PORT},${PREFIX}4:${PORT}"
TAG        = cs128-hw3
NAME       = --name REPLICA_
REPLICA2   = ${NET} ${IP}2 ${VIEW} -e IP_PORT="${PREFIX}2:${PORT}" ${NAME}2 ${TAG}
REPLICA3   = ${NET} ${IP}3 ${VIEW} -e IP_PORT="${PREFIX}3:${PORT}" ${NAME}3 ${TAG}
REPLICA4   = ${NET} ${IP}4 ${VIEW} -e IP_PORT="${PREFIX}4:${PORT}" ${NAME}4 ${TAG}
SINGLE     = ${NET} ${IP}2 -e IP_PORT="${PREFIX}2:${PORT}" ${NAME}1 ${TAG}

RUNNING   := $(shell docker ps | grep REPLICA) # DELETE
STOPPED   := $(shell docker ps -a | grep REPLICA) # DELETE
IMAGES    := $(shell docker images -a | grep REPLICA | awk ' $1 { print $1 } ') # DELETE


# This will run the docker build command, which builds the app
default : docker                            # DELETE

# Everything executes the build, unit tests, and runs the integration test
all : docker                                # DELETE
	${UNIT}                                 # DELETE
	${TEST}                                 # DELETE

# This runs 'go build ...'
app :
	${BUILD} -o ${EXEC} ${LD} ${SOURCES}

# This runs the unit tests
unit :
	${UNIT}

# This runs the integration tests
test : docker                               # DELETE
	${TEST}                                 # DELETE

# This removes the binary
spotless :
ifneq (, $(shell ls | grep ${EXEC}))
	- rm ${EXEC}
endif

# Alias for above
clean : spotless

# Rebuild the thing
again :
	${MAKE} spotless all

# This runs the Docker build command
docker : network cleanfile                     # DELETE
	${DOCKERBLD} ${DFLAGS}                     # DELETE

# This makes a copy of the Makefile that doesn't try to run 
# Docker so we don't get a warning about it not being available.
cleanfile :
	awk '!/DELETE/' Makefile > Makefile.docker


# This builds the subnet in Docker
network :    # DELETE
ifeq (, $(shell docker network ls | grep mynet)) # DELETE
	- sudo docker network create --subnet=10.0.0.0/16 mynet # DELETE
endif # DELETE

# This runs a single container detached
run : dockerclean docker                       # DELETE
	${DOCKERRUN} ${SINGLE}                     # DELETE

# This builds a set of three nodes that should be able to talk to each other
cluster : dockerclean docker                   # DELETE
	${DOCKERRUN} ${REPLICA2}                   # DELETE
	${DOCKERRUN} ${REPLICA3}                   # DELETE
	${DOCKERRUN} ${REPLICA4}                   # DELETE

dockerclean :                                                                                                # DELETE
	$(shell docker kill $$(docker ps -a | grep REPLICA | awk ' $$1 { print $$1 } ') > /dev/null 2>&1)        # DELETE
	$(shell docker rm $$(docker ps -a | grep REPLICA | awk ' $$1 { print $$1 } ') > /dev/null 2>&1)          # DELETE
	$(shell docker rmi $$(docker images -a | grep cs128 | awk ' $$1 { print $$1 } ') > /dev/null 2>&1)       # DELETE