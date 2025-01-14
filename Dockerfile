FROM golang:1.23-alpine3.21 AS builder

COPY ./ /app
WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on

RUN go build -a -installsuffix cgo -ldflags '-s -w' -o exapp-go *.go

FROM  alpine:3.21

COPY --from=builder /app/exapp-go /

ADD config /config


CMD ["/exapp-go"]
