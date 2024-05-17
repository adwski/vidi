FROM golang:1.22.3-bookworm as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOPATH=/go

ADD . /build

WORKDIR /build

RUN go mod download
RUN <<EOF
    go build -o uploader -ldflags '-d -w -s' -tags netgo -installsuffix netgo ./cmd/uploader/*.go
    chmod +x /build/uploader
EOF


FROM builder as dev

ENTRYPOINT ["/build/uploader"]


FROM gcr.io/distroless/static as release

WORKDIR /
USER nonroot:nonroot
COPY --from=builder /build/uploader /uploader

ENTRYPOINT ["/uploader"]
