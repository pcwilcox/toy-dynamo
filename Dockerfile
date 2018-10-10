# Build the binary from alpine
FROM golang:alpine as builder
RUN mkdir /app
COPY . /app/
WORKDIR /app

# Alpine images don't have git installed so we need it
# in order to download our dependencies.
RUN apk add --no-cache git mercurial \
    && go get -d -v

# Build the app.
RUN go build -o cs128-hw1 .

# Expose the required port
EXPOSE 8080
CMD ["/app/cs128-hw1"]
