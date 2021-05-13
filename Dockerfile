# Builder container
FROM docker.io/library/golang:1.16.3-alpine AS builder
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build

# App container
FROM alpine
COPY --from=builder /go/src/app/chat_v3 /bin/
ENTRYPOINT ["/bin/chat_v3"]