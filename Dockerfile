# Build the binary from alpine
FROM golang:alpine
RUN mkdir /app
COPY . /app/
WORKDIR /app

# Alpine images don't have git installed so we need it
# in order to download our dependencies.
RUN apk add --no-cache --update \
    git mercurial make \
    && go get -d -v

# Build the app.
RUN make

# Expose the required port
EXPOSE 8080
CMD ["/app/cs128-hw2"]
