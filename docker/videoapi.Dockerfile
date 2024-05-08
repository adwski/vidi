FROM golang:1.22.1-bookworm as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOPATH=/go

ADD . /build

WORKDIR /build

RUN go mod download
RUN <<EOF
    go build -o videoapi -ldflags '-d -w -s' -tags netgo -installsuffix netgo ./cmd/videoapi/*.go
    chmod +x /build/videoapi
EOF


FROM builder as dev

ENTRYPOINT ["/build/videoapi"]


FROM gcr.io/distroless/static as release

WORKDIR /
USER nonroot:nonroot
COPY --from=builder /build/videoapi /videoapi

ENTRYPOINT ["/videoapi"]
