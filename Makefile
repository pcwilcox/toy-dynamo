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
COVERFILE = out

# Change this to the application name
EXEC      = cs128-hw2

# Add source files to this list
SOURCES   = main.go dbAccess.go app.go kvs.go

all : ${EXEC}

${EXEC} :
	${BUILD} -o ${EXEC} ${SOURCES}

test :
	${TEST}

bench :
	${BENCH}

benchmark : bench

cover :
	${COVER} && rm ${COVERFILE}

spotless :
ifneq (, $(shell ls | grep ${EXEC}))
	- rm ${EXEC}
endif

again :
	${MAKE} spotless all