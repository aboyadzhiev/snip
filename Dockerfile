FROM golang:1.24.0 AS build-stage

WORKDIR /go/src/snip

COPY . .

RUN go mod download
RUN go vet ./...
RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s" -o /go/bin/snip cmd/main.go

FROM gcr.io/distroless/static-debian12 AS release-stage

WORKDIR /

COPY --from=build-stage /go/bin/snip /usr/local/bin/snip

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/snip"]