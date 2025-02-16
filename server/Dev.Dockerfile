FROM golang:1.24.0

WORKDIR /go/src/snip

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s" -o /usr/local/bin/snip cmd/main.go

ENTRYPOINT ["/usr/local/bin/snip"]
