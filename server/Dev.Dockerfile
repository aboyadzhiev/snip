FROM golang:1.24.0

WORKDIR /go/src/snip

COPY . .

RUN go mod download

