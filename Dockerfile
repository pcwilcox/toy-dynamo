# Build the binary from alpine
FROM golang:alpine
RUN mkdir /app
COPY . /app/
WORKDIR /app

# Alpine images don't have git or make installed 
# so we'll install them and then get the
# app dependencies
RUN apk add --no-cache --update \
    git mercurial make \
    && go get -d -v

# Build the app.
RUN make

# Expose the required port
EXPOSE 8080

# Run the sucker
CMD ["/app/cs128-hw3"]
