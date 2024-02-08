FROM golang:1.21.4-bookworm as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOPATH=/go

ADD . /build

WORKDIR /build

RUN go mod download
RUN <<EOF
    go build -o streamer -ldflags '-d -w -s' -tags netgo -installsuffix netgo ./cmd/streamer/*.go
    chmod +x /build/streamer
EOF


FROM builder as dev

ENTRYPOINT ["/build/streamer"]


FROM gcr.io/distroless/static as release

WORKDIR /
USER nonroot:nonroot
COPY --from=builder /build/streamer /streamer

ENTRYPOINT ["/streamer"]
