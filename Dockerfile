FROM golang:onbuild
RUN mkdir /app
COPY . /app/
WORKDIR /app
RUN go build -o restful .
EXPOSE 8080
CMD ["/app/restful"]