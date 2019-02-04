FROM golang:alpine
WORKDIR $GOPATH/src/github.com/ageapps/gambercoin
ADD . .
RUN go get ./...
RUN go build -o node_server ./cmd/node_server
RUN adduser -S -D -H -h /app appuser
USER appuser
ENV SERVER_PORT=8080
CMD ["./node_server"]
