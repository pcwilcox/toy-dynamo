FROM golang:onbuild
RUN mkdir /app
COPY . /app/
WORKDIR /app
RUN go test
RUN go build -o cs128-hw1 .
EXPOSE 8080
CMD ["/app/cs128-hw1"]