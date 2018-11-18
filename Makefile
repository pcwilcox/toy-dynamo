# Makefile
# 
# CMPS 128, Fall 2018
#
# Lawrence Lawson    lelawson
# Pete Wilcox        pcwilcox
#
# Incorporates compile-time variables into the build

# Predefined stuff
BUILD     = go build
TEST      = go test
BENCH     = ${TEST} -bench=.
COVER     = ${TEST} -coverprofile ${COVERFILE}
DOCKER    = docker build
COVERFILE = out
DFLAGS    = -t ${EXEC} . 

# Grabs the name of the current branch
BRANCH   := $(shell git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/\1/')

# Grabs the total number of commits
BUILDNUM := $(shell git rev-list --count HEAD) 

# Grabs the last chunk of the current commit ID
HASH     := $(shell git rev-parse --short HEAD 2> /dev/null)

# Change this to the application name
EXEC      = cs128-hw3

# Add source files to this list
SOURCES   = main.go dbAccess.go app.go kvs.go restful.go values.go view.go

# Flags passed to the linker which define version strings in main.go
LDFLAGS   = '-X "main.branch=${BRANCH}" -X "main.hash=${HASH}" -X "main.build=${BUILDNUM}"'
LD        = -ldflags ${LDFLAGS}

# Everything executes the build
all : ${EXEC}

# This runs 'go build ...'
${EXEC} :
	${BUILD} -o ${EXEC} ${LD} ${SOURCES} 

# This runs the integration test script
test : docker
	${TEST}

# This runs Go benchmarks (need to be written)
bench :
	${BENCH}

# Alias for above
benchmark : bench

# This runs the Go coverage tool
cover :
	${COVER} && rm ${COVERFILE}

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
docker : network
	${DOCKER} ${DFLAGS}

# This builds the subnet in Docker
network :
ifeq (, $(shell docker network ls | grep mynet))
	- sudo docker network create --subnet=10.0.0.0/16 mynet
endif

leader : docker
	docker run -p 8083:8080 --net=mynet --ip=10.0.0.20 -d ${EXEC}

follower : docker
	docker run -p 8084:8080 --net=mynet -e MAINIP=10.0.0.20:8080 -d ${EXEC}

