# Dockerfile
#
# CMPS 128 Fall 2018
#
# Lawrence Lawson     lelawson
# Pete Wilcox         pcwilcox
# Annie Shen          ashen7
# Victoria Tran       vilatran
#
# This file defines the multi-stage build process we use to pull 
# dependencies, compile our code, and build a container.
#

# STEP ONE:
# The first step is to build the binary. We use golang:alpine as
# it is very small and is able to compile golang. We copy our
# source to the build layer and run make, which executes our
# build command.
FROM golang:alpine AS builder
RUN mkdir /app
COPY . /app/
WORKDIR /app

# Alpine images don't have git or make installed, need them 
RUN apk add --no-cache --update git mercurial make 

# This command pulls the libraries we use    
RUN go get -d -v

# Build the app - this runs a customized build command specified
# in the Makefile in order to strip debug info, embed our version
# info, and compile for a Linux platform.
RUN make -f Makefile.docker app

# STEP TWO:
# Use the smallest possible container to run the binary
FROM scratch

# Copy our static executable from the build layer
COPY --from=builder /app/app /app

# Expose the required port
EXPOSE 8080

# Run the sucker
ENTRYPOINT ["/app"]
