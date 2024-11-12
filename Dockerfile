FROM golang:1.22-alpine3.18 AS builder

COPY ./ /app
WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on

RUN go build -a -installsuffix cgo -ldflags '-s -w' -o exapp-go *.go

FROM  alpine:3.17

COPY --from=builder /app/exapp-go /

ADD config /config


CMD [/exapp-go]
