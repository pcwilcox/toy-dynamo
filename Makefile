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

# Change this to the application name
EXEC      = cs128-hw2

# Add source files to this list
SOURCES   = main.go dbAccess.go app.go kvs.go restful.go forward.go values.go

# Everything executes the build
all : ${EXEC}

# This runs 'go build ...'
${EXEC} :
	${BUILD} -o ${EXEC} ${SOURCES}

# This runs 'go test ...'
test :
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
docker :
	${DOCKER} ${DFLAGS}

# This builds the subnet in Docker
network :
	sudo docker network create --subnet=10.0.0.0/16 mynet

leader :
	docker run -p 8083:8080 --net=mynet --ip=10.0.0.20 -d ${EXEC}

follower :
	docker run -p 8084:8080 --net=mynet -e MAINIP=10.0.0.20:8080 -d ${EXEC}

